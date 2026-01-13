package llm

import (
	"context"
	"github.com/leon37/FaceTaxLedger/internal/model"
)

// Provider 定义了 LLM 的通用行为
type Provider interface {
	// AnalyzeExpense 接收用户输入，返回结构化的分析结果
	AnalyzeExpense(ctx context.Context, userContext string) (*model.FaceTaxAnalysis, error)
}
