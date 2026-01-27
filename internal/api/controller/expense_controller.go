package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/leon37/FaceTaxLedger/internal/api/response"
	"github.com/leon37/FaceTaxLedger/internal/model"
	"github.com/leon37/FaceTaxLedger/internal/repository"
	"github.com/leon37/FaceTaxLedger/internal/service"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type ExpenseController struct {
	service *service.ExpenseService // 依赖 Service
}

// NewExpenseController 构造函数
func NewExpenseController(s *service.ExpenseService) *ExpenseController {
	return &ExpenseController{service: s}
}

// ExpenseAnalyzeRequest 定义前端传来的 JSON 参数结构
type ExpenseAnalyzeRequest struct {
	Description string `json:"description" binding:"required"`
}

type ExpenseAnalyzeResponse struct {
	Id       string  `json:"id"`
	Comment  string  `json:"comment"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Date     string  `json:"date"`
	Note     string  `json:"note"`
}

// Analyze 智能记账
// @Summary 自然语言记账
// @Description AI 自动提取金额、分类并生成吐槽。
// @Tags Expense
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ExpenseAnalyzeRequest true "记账内容"
// @Success 200 {object} response.Response{data=controller.ExpenseAnalyzeResponse}
// @Router /expenses/analyze [post]
func (ctrl *ExpenseController) Analyze(c *gin.Context) {
	userIDStr := c.GetString("userID")
	if userIDStr == "" {
		response.Error(c, http.StatusUnauthorized, "缺少 X-User-ID 请求头")
		return
	}

	// 2. 解析 JSON 参数
	var req ExpenseAnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 参数校验失败
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	// 2. 设置 SSE Header
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	slog.Info("收到 API 记账请求", "description", req.Description)

	// 3. 调用 Service 业务逻辑
	// 注意：Context 应该传递下去，用于链路追踪或超时控制
	ei := service.ExpenseInput{
		UserID:      userIDStr,
		Description: req.Description,
	}
	streamCh, commitFunc, err := ctrl.service.StreamExpense(c.Request.Context(), ei)
	if err != nil {
		slog.Error("API 调用业务层失败", "error", err)
		response.Error(c, http.StatusInternalServerError, "AI 大脑短路了，请稍后再试")
		return
	}
	// 4. 循环读取流，推送到前端，同时在内存拼接
	var fullJSONBuilder strings.Builder

	// 监听客户端断开 (Gin 的特性)
	clientGone := c.Writer.CloseNotify()
	for {
		select {
		case <-clientGone:
			// 客户端断开了，停止处理
			return
		case fragment, ok := <-streamCh:
			if !ok {
				// 通道关闭，说明流结束了 -> 进入结算阶段
				goto Finalize
			}

			// A. 推送给前端 (Raw Fragment)
			// 前端收到后拼接到 buffer 中尝试解析
			c.SSEvent("delta", fragment)

			// B. 后端累积
			fullJSONBuilder.WriteString(fragment)

			// 这里的 Flush 很重要，确保数据立刻发出去
			c.Writer.Flush()
		}
	}

Finalize:
	// 5. 流传输完毕，执行落库逻辑
	fullJSON := fullJSONBuilder.String()
	expense, err := commitFunc(fullJSON)
	if err != nil {
		c.SSEvent("error", "saving failed: "+err.Error())
		return
	}

	finalData, _ := json.Marshal(expense)
	c.SSEvent("done", string(finalData))
}

// ListRequest 列表请求参数
type ListRequest struct {
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"page_size,default=10"`
	Category  string `form:"category"`
	StartDate string `form:"start_date"` // 格式 2023-01-01
	EndDate   string `form:"end_date"`
}

type ListResponse struct {
	List  []model.ExpenseEntity `json:"list"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
}

// List 智能记账
// @Summary 获取记账列表
// @Description 按条件筛选账单
// @Tags Expense
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ListRequest true "筛选条件"
// @Success 200 {object} response.Response{data=controller.ListResponse} "注意：这里加上了 controller. 前缀"
// @Router /expenses [get]
func (ctrl *ExpenseController) List(c *gin.Context) {
	// 1. 获取 UserID (Header)
	userIDStr := c.GetString("userID")

	// 2. 绑定 Query 参数
	var req ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	// 3. 构造 Filter
	filter := repository.ExpenseFilter{
		UserID:   userIDStr,
		Category: req.Category,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	// 解析时间字符串 (简单处理)
	if req.StartDate != "" {
		t, _ := time.Parse("2006-01-02", req.StartDate)
		filter.StartDate = t
	}
	if req.EndDate != "" {
		t, _ := time.Parse("2006-01-02", req.EndDate)
		filter.EndDate = t.Add(24 * time.Hour) // 包含当天
	}

	// 4. 调用 Service
	list, total, err := ctrl.service.GetExpensesList(c.Request.Context(), filter)
	if err != nil {
		slog.Error("获取账单列表失败", "error", err)
		response.Error(c, http.StatusInternalServerError, "获取列表失败")
		return
	}
	rsp := ListResponse{
		List:  list,
		Total: total,
		Page:  req.Page,
	}

	// 5. 返回带分页信息的响应
	response.Success(c, rsp)
}

type DeleteRequest struct {
	ID int64 `json:"id" binding:"required"`
}

// Delete 删除账本条目
// @Summary 删除账本条目
// @Description 删除已存在的账单信息，仅限本人操作
// @Tags Expense
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DeleteRequest true "删除参数"
// @Success 200 {object} response.Response "成功"
// @Router /expenses/delete [post]
func (ctrl *ExpenseController) Delete(c *gin.Context) {
	userIDStr := c.GetString("userID")

	// 获取 URL 中的 ID 参数
	var req DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := ctrl.service.DeleteExpense(c.Request.Context(), userIDStr, req.ID); err != nil {
		// 这里可以细分错误类型，比如“无权操作”返回 403
		slog.Error("删除失败", "id", req.ID, "error", err)
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil)
}

type UpdateRequest struct {
	ID       int64   `json:"id" binding:"required"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Date     string  `json:"date"`
	Note     string  `json:"note"`
}

// Update 更新账本条目
// @Summary 更新账本条目
// @Description 修改已存在的账单信息，仅限本人操作
// @Tags Expense
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateRequest true "更新参数"
// @Success 200 {object} response.Response "成功"
// @Router /expenses/update [post]
func (ctrl *ExpenseController) Update(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		slog.Error("鉴权失败")
		response.Error(c, http.StatusUnauthorized, "User ID not found in context")
		return
	}
	userID := val.(string)

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := ctrl.service.UpdateExpense(c.Request.Context(), userID, req.ID, req.Category, req.Amount, req.Note); err != nil {
		// 这里可以细分错误类型，比如“无权操作”返回 403
		slog.Error("更新失败", "id", req.ID, "error", err)
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil)
}
