package repository

import (
	"context"
	// 这里会引用到你前面写的 internal/infrastructure/vectordb 里的结构
)

// MemoryRepo 定义了 AI 记忆相关的接口
type MemoryRepo interface {
	SaveMemory(ctx context.Context, uuid string, expenseID uint, description string) error
	SearchSimilar(ctx context.Context, uuid string, description string, limit int) ([]string, error)
}
