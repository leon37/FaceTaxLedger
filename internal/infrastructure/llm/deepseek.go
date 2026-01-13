package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/leon37/FaceTaxLedger/internal/model"
	"io"
	"net/http"
	"strings"
	"time"
)

type DeepSeekClient struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
	modelName  string
}

// OpenAI 兼容的请求结构
type chatRequest struct {
	Model          string    `json:"model"`
	Messages       []message `json:"messages"`
	ResponseFormat format    `json:"response_format"` // 关键：强制 JSON 模式
	Temperature    float64   `json:"temperature"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type format struct {
	Type string `json:"type"`
}

// OpenAI 兼容的响应结构
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewDeepSeekClient(apiKey string) *DeepSeekClient {
	return &DeepSeekClient{
		apiKey:    apiKey,
		apiURL:    "https://api.deepseek.com/chat/completions", // DeepSeek V3/R1 官方端点
		modelName: "deepseek-chat",                             // 或者 deepseek-reasoner
		httpClient: &http.Client{
			Timeout: 30 * time.Second, // 务必设置超时，防止 Goroutine 泄漏
		},
	}
}

func (c *DeepSeekClient) AnalyzeExpense(ctx context.Context, userInput string) (*model.FaceTaxAnalysis, error) {
	// 1. 构建请求 Payload
	reqBody := chatRequest{
		Model: c.modelName,
		Messages: []message{
			{Role: "system", Content: model.SystemPrompt}, // 注入我们在 Context 中定义的 Prompt
			{Role: "user", Content: userInput},
		},
		// 关键点：启用 JSON Mode，这能极大提高模型输出 JSON 的稳定性
		ResponseFormat: format{Type: "json_object"},
		Temperature:    1.3, // 稍微高一点，让毒舌更有创意
	}

	jsonBody, _ := json.Marshal(reqBody)

	// 2. 创建 HTTP Request
	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 3. 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api call failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	// 4. 处理 HTTP 错误状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// 5. 解析外层响应
	var apiResp chatResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse api response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, errors.New("api returned no choices")
	}

	rawContent := apiResp.Choices[0].Message.Content

	// 6. 清洗数据 (防守型编程)
	// 虽然开了 json_object，但有时候模型可能会带上 ```json ... ``` 的 Markdown 标记
	cleanContent := sanitizeJSON(rawContent)

	// 7. 反序列化为领域模型
	var analysis model.FaceTaxAnalysis
	if err := json.Unmarshal([]byte(cleanContent), &analysis); err != nil {
		// Log raw content for debugging here if necessary
		return nil, fmt.Errorf("failed to parse domain model: %w | raw: %s", err, cleanContent)
	}

	return &analysis, nil
}

// sanitizeJSON 去除可能的 Markdown 标记
func sanitizeJSON(input string) string {
	input = strings.TrimSpace(input)
	input = strings.TrimPrefix(input, "```json")
	input = strings.TrimPrefix(input, "```")
	input = strings.TrimSuffix(input, "```")
	return input
}
