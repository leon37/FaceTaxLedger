package repository

import (
	"context"
	// 这里会引用到你前面写的 internal/infrastructure/vectordb 里的结构
)

type MemoryResult struct {
	Content   string
	Category  string
	Timestamp int64
}

// MemoryRepo 定义了 AI 记忆相关的接口
type MemoryRepo interface {
	SaveMemory(ctx context.Context, uuid string, expenseID uint, description string, category string, vector []float32) error
	SearchSimilar(ctx context.Context, uuid string, limit int, queryVector []float32) ([]MemoryResult, error)
	Delete(ctx context.Context, id int64) error
}
