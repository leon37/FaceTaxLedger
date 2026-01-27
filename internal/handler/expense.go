package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leon37/FaceTaxLedger/internal/service"
)

// ExpenseHandler 持有 Service 的依赖
type ExpenseHandler struct {
	svc *service.ExpenseService
}

// NewExpenseHandler 构造函数
func NewExpenseHandler(svc *service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{
		svc: svc,
	}
}

// HandleSubmitExpense 处理 POST /api/v1/expenses
func (h *ExpenseHandler) HandleSubmitExpense(c *gin.Context) {
	// 1. 绑定参数 (Binding)
	// 这里直接复用 Service 层的 Input 结构，或者单独定义 DTO 都可以
	// 为了省事，我们先复用 Service 的定义
	var input service.ExpenseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误: " + err.Error()})
		return
	}

	// 2. 调用业务逻辑
	// Context 必须往下传，以便处理超时或取消
	_, _, err := h.svc.StreamExpense(c.Request.Context(), input)
	if err != nil {
		// 这里可以根据 error 类型细分状态码，比如 500 还是 400
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. 返回响应
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "记账成功",
		"data": result,
	})
}
