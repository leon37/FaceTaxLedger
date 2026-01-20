package service

import (
	"context"
	"fmt"
	"github.com/leon37/FaceTaxLedger/internal/infrastructure/embedding"
	"github.com/leon37/FaceTaxLedger/internal/repository"
	"log/slog"
	"time"

	"github.com/leon37/FaceTaxLedger/internal/infrastructure/llm"
	"github.com/leon37/FaceTaxLedger/internal/model"
)

// ExpenseInput 是前端传来的原始参数 (DTO)
type ExpenseInput struct {
	UserID      string `json:"user_id"`
	Description string `json:"description"` // 例如："请客吃饭"
}

// ExpenseResult 是返回给前端的完整结果 (VO)
type ExpenseResult struct {
	ExpenseID string                 `json:"expense_id"`
	Analysis  *model.FaceTaxAnalysis `json:"analysis"` // 包含毒舌评论
	SavedAt   time.Time              `json:"saved_at"`
}

// ExpenseService 定义业务逻辑接口
type ExpenseService struct {
	llmClient  llm.Provider // 依赖接口，而不是具体 struct！(关键点)
	embedder   embedding.Provider
	repo       repository.ExpenseRepo // 稍后我们会注入数据库仓储
	memoryRepo repository.MemoryRepo
}

// NewExpenseService 构造函数 (依赖注入)
func NewExpenseService(llmClient llm.Provider, embedder embedding.Provider, repo repository.ExpenseRepo, memory repository.MemoryRepo) *ExpenseService {
	return &ExpenseService{
		llmClient:  llmClient,
		embedder:   embedder,
		repo:       repo,
		memoryRepo: memory,
	}
}

// SubmitExpense 处理一次完整的记账请求
func (s *ExpenseService) SubmitExpense(ctx context.Context, input ExpenseInput) (*ExpenseResult, error) {
	slog.Info("收到记账请求",
		"uid", input.UserID,
		"description", input.Description)
	// 1. RAG 检索：先查历史 (比如查最近相似的 3 条)
	// 这一步不能报错阻断流程，如果检索失败，就当没有历史
	var historyContext []repository.MemoryResult
	queryVector, err := s.embedder.GetVector(ctx, input.Description)
	if err != nil {
		slog.Error("Embed failed: %v\n", err)
		return nil, err
	}
	if similarLogs, err := s.memoryRepo.SearchSimilar(ctx, input.UserID, 3, queryVector); err == nil {
		historyContext = similarLogs
	} else {
		// 记录日志但不报错
		slog.Error("RAG Search failed: %v\n", err)
		return nil, err
	}

	prompt := fmt.Sprintf("用户当前消费描述: %s", input.Description)

	if len(historyContext) > 0 {
		prompt += fmt.Sprintf("\n\n【用户相关历史消费（参考这些黑历史来吐槽）】：\n")
		for i, h := range historyContext {
			timeStr := formatTimeAgo(h.Timestamp)
			prompt += fmt.Sprintf("%d. [%s] %s\n", i+1, timeStr, h.Content)
		}
		prompt += "\n请结合历史行为，如果发现用户在短时间内（如几天内）重复此类消费，请加大力度嘲讽。"
	}

	analysis, err := s.llmClient.AnalyzeExpense(ctx, prompt)
	if err != nil {
		return nil, err
	}

	expenseTime, err := time.Parse("2006-01-02 15:04:05", analysis.Date)
	if err != nil {
		expenseTime = time.Now() // 解析失败就兜底用当前时间
	}

	entity := &model.ExpenseEntity{
		UserID:    input.UserID,
		Amount:    analysis.Amount,
		Comment:   analysis.Comment,
		Category:  analysis.Category,
		CreatedAt: expenseTime,
		Note:      analysis.Note,
	}
	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, err
	}

	// 5. 【异步】存入记忆 (Fire and Forget)
	// 不要让用户等待 Embedding 的过程，开个协程去存
	go func() {
		// 创建一个新的 context，因为外面的 ctx 可能会在请求结束时取消
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		vector, err := s.embedder.GetVector(bgCtx, input.Description)
		if err != nil {
			slog.Error("Failed to embed vector", "error", err)
		}
		if err := s.memoryRepo.SaveMemory(bgCtx, input.UserID, entity.ID, input.Description, vector); err != nil {
			slog.Error("Failed to save memory", "error", err)
		}
	}()

	return &ExpenseResult{
		ExpenseID: fmt.Sprintf("%d", entity.ID), // ID 已经是数据库自增生成的了
		Analysis:  analysis,
		SavedAt:   entity.CreatedAt,
	}, nil
}

func formatTimeAgo(timestamp int64) string {
	if timestamp == 0 {
		return "很久以前"
	}
	delta := time.Since(time.Unix(timestamp, 0))
	hours := delta.Hours()

	if hours < 24 {
		return "今天"
	} else if hours < 48 {
		return "昨天"
	} else {
		days := int(hours / 24)
		return fmt.Sprintf("%d天前", days)
	}
}

// GetExpensesList 获取列表
func (s *ExpenseService) GetExpensesList(ctx context.Context, filter repository.ExpenseFilter) ([]model.ExpenseEntity, int64, error) {
	// 这里可以加一些额外的业务逻辑，比如数据脱敏等，目前直接透传
	return s.repo.List(ctx, filter)
}

// DeleteExpense 删除账单 (带归属权校验)
func (s *ExpenseService) DeleteExpense(ctx context.Context, userID string, expenseID int64) error {
	// 1. 先查出来，确认是否存在
	existing, err := s.repo.GetByID(ctx, expenseID)
	if err != nil {
		return fmt.Errorf("账单不存在或查询失败: %w", err)
	}

	// 2. 🛡️ 安全核心：检查这条账单是不是这个人的
	if existing.UserID != userID {
		return fmt.Errorf("无权操作此账单")
	}

	// 3. 执行删除 (MySQL)
	// 思考：是否要同步删除 Qdrant 里的记忆？
	// 这是一个复杂的分布式一致性问题。简单起见，目前只删账本。
	// 如果不删 Qdrant，AI 可能会记得“你花过”，但账本里没记录，这通常可以接受（当作由于某种原因没记账）。
	if err := s.repo.Delete(ctx, expenseID); err != nil {
		return err
	}

	go func() {
		if err := s.memoryRepo.Delete(context.Background(), expenseID); err != nil {
			slog.Error("Qdrant 删除记忆失败", "id", expenseID, "error", err)
		} else {
			slog.Info("Qdrant 记忆已同步删除", "id", expenseID)
		}
	}()
	return nil
}

// UpdateExpense 更新账单
func (s *ExpenseService) UpdateExpense(ctx context.Context, userID string, expenseID int64, category string, amount float64, note string) error {
	existing, err := s.repo.GetByID(ctx, expenseID)
	if err != nil {
		return err
	}

	if existing.UserID != userID {
		return fmt.Errorf("无权操作此账单")
	}

	// 更新字段
	if len(category) > 0 {
		existing.Category = category
	}
	if amount > 0 {
		existing.Amount = amount
	}
	if len(note) > 0 {
		existing.Note = note
	}
	// 注意：修改账单通常不会重新触发 AI 分析，除非你希望这样设计

	err = s.repo.Update(ctx, existing)
	if err != nil {
		return err
	}

	go func() {
		// 1. 重新生成文本
		newContent := fmt.Sprintf("消费: %s, 金额: %.2f, 备注: %s", category, amount, note)

		// 2. 重新 Embedding (这一步可能耗时，所以放协程)
		vec, err := s.embedder.GetVector(context.Background(), newContent)
		if err != nil {
			slog.Error("更新记忆时生成向量失败", "error", err)
			return
		}

		// 3. 覆盖保存 (Qdrant 的 Upsert 会自动覆盖旧数据)
		if err := s.memoryRepo.SaveMemory(context.Background(), userID, uint(expenseID), newContent, vec); err != nil {
			slog.Error("Qdrant 更新记忆失败", "error", err)
		} else {
			slog.Info("Qdrant 记忆已同步更新", "id", expenseID)
		}
	}()

	return nil
}
