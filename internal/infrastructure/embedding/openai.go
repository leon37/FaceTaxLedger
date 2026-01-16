package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OpenAIClient struct {
	apiKey  string
	baseURL string // 例如 "https://api.openai.com/v1" 或其它中转地址
	model   string // 例如 "text-embedding-3-small"
}

func NewOpenAIClient(apiKey, baseURL, model string) *OpenAIClient {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "text-embedding-3-small" // 默认模型，维度 1536
	}
	return &OpenAIClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
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
	reqBody := embeddingRequest{
		Input: text,
		Model: c.model,
	}
	jsonData, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/embeddings", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("embedding api error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp embeddingResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("empty embedding data returned")
	}

	return apiResp.Data[0].Embedding, nil
}
