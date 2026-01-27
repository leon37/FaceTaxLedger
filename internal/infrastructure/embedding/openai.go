package embedding

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	model  string // 例如 "text-embedding-3-small"
	client *openai.Client
}

func NewOpenAIClient(apiKey, baseURL, model string) *OpenAIClient {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "text-embedding-3-small" // 默认模型，维度 1536
	}
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	return &OpenAIClient{
		model:  model,
		client: openai.NewClientWithConfig(config),
	}
}

// 请求结构体
type embeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// 响应结构体
type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *OpenAIClient) GetVector(ctx context.Context, text string) ([]float32, error) {
	req := openai.EmbeddingRequest{
		Input: []string{text},
		// 如果 c.model 是字符串，可以强转，或者直接使用 openai.SmallEmbedding3 等常量
		Model: openai.EmbeddingModel(c.model),
	}

	// 发起请求
	resp, err := c.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openai embedding failed: %w", err)
	}

	// 校验返回数据
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("empty embedding data returned")
	}

	// 返回第一个结果的向量
	return resp.Data[0].Embedding, nil
}
