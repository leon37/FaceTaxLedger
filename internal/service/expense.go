package service

import (
	"context"
	"fmt"
	"github.com/leon37/FaceTaxLedger/internal/repository"
	"time"

	"github.com/leon37/FaceTaxLedger/internal/infrastructure/llm"
	"github.com/leon37/FaceTaxLedger/internal/model"
)

// ExpenseInput 是前端传来的原始参数 (DTO)
type ExpenseInput struct {
	UserID      string  `json:"user_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"` // 例如："请客吃饭"
}

// ExpenseResult 是返回给前端的完整结果 (VO)
type ExpenseResult struct {
	ExpenseID string                 `json:"expense_id"`
	Analysis  *model.FaceTaxAnalysis `json:"analysis"` // 包含毒舌评论
	SavedAt   time.Time              `json:"saved_at"`
}

// ExpenseService 定义业务逻辑接口
type ExpenseService struct {
	llmClient  llm.Provider           // 依赖接口，而不是具体 struct！(关键点)
	repo       repository.ExpenseRepo // 稍后我们会注入数据库仓储
	memoryRepo repository.MemoryRepo
}

// NewExpenseService 构造函数 (依赖注入)
func NewExpenseService(llmClient llm.Provider, repo repository.ExpenseRepo, memory repository.MemoryRepo) *ExpenseService {
	return &ExpenseService{
		llmClient:  llmClient,
		repo:       repo,
		memoryRepo: memory,
	}
}

// SubmitExpense 处理一次完整的记账请求
func (s *ExpenseService) SubmitExpense(ctx context.Context, input ExpenseInput) (*ExpenseResult, error) {
	// 1. RAG 检索：先查历史 (比如查最近相似的 3 条)
	// 这一步不能报错阻断流程，如果检索失败，就当没有历史
	var historyContext []string
	if similarLogs, err := s.memoryRepo.SearchSimilar(ctx, input.UserID, input.Description, 3); err == nil {
		historyContext = similarLogs
	} else {
		// 记录日志但不报错
		fmt.Printf("RAG Search failed: %v\n", err)
	}

	prompt := fmt.Sprintf("用户消费了 %.2f 元，备注：%s。", input.Amount, input.Description)

	if len(historyContext) > 0 {
		prompt += fmt.Sprintf("\n\n【用户相关历史消费（参考这些黑历史来吐槽）】：\n")
		for i, h := range historyContext {
			prompt += fmt.Sprintf("%d. %s\n", i+1, h)
		}
		prompt += "\n请结合历史行为，如果发现他在重复犯错（比如一直喝奶茶），请加大力度嘲讽。"
	}

	analysis, err := s.llmClient.AnalyzeExpense(ctx, prompt)
	if err != nil {
		return nil, err
	}

	entity := &model.ExpenseEntity{
		UserID:       input.UserID,
		Amount:       input.Amount,
		Description:  input.Description,
		IsFaceTax:    analysis.IsFaceTax,
		TaxCategory:  analysis.TaxCategory,
		Comment:      analysis.Comment,
		SarcasmLevel: analysis.SarcasmLevel,
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

		if err := s.memoryRepo.SaveMemory(bgCtx, input.UserID, entity.ID, input.Description); err != nil {
			fmt.Printf("Failed to save memory: %v\n", err)
		}
	}()

	return &ExpenseResult{
		ExpenseID: fmt.Sprintf("%d", entity.ID), // ID 已经是数据库自增生成的了
		Analysis:  analysis,
		SavedAt:   entity.CreatedAt,
	}, nil
}
