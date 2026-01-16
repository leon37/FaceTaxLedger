package embedding

import "context"

// Provider 定义了将文本转换为向量的能力
type Provider interface {
	// GetVector 输入文本，返回 float32 数组
	GetVector(ctx context.Context, text string) ([]float32, error)
}
