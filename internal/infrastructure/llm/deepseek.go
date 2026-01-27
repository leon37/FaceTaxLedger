package llm

import (
	"context"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io"
	"log/slog"
	"time"
)

type DeepSeekClient struct {
	modelName string
	client    *openai.Client
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

func NewDeepSeekClient(apiKey, baseUrl, modelName string) *DeepSeekClient {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseUrl

	return &DeepSeekClient{
		modelName: modelName,
		client:    openai.NewClientWithConfig(config),
	}
}

func (d *DeepSeekClient) AnalyzeExpense(ctx context.Context, userContext string, categories []string, historyContext []string, enableRoast bool) (<-chan string, error) {
	// 1. 构建 System Prompt
	sysPrompt := fmt.Sprintf("你是一个专业的记账助手。当前系统时间：%s。", time.Now().Format("2006-01-02 15:04:05"))
	var contextInstruction string

	if len(historyContext) > 0 {
		// 拼接历史记录字符串
		historyStr := "\n\n【用户相关历史消费参考】:\n"
		for _, log := range historyContext {
			historyStr += fmt.Sprintf("- %s\n", log)
		}

		if enableRoast {
			// === 场景 A: 开启吐槽 ===
			// 指令：用历史数据来攻击
			contextInstruction = historyStr + "\n请结合上述历史行为，如果发现用户在短时间内重复消费或有不良消费习惯，请在 comment 字段中加大力度进行辛辣、幽默的吐槽。"
		} else {
			// === 场景 B: 关闭吐槽 (纯记账模式) ===
			// 指令：用历史数据来校准分类
			contextInstruction = historyStr + "\n请参考上述历史消费的'分类'和'备注'习惯。如果当前消费与历史记录相似，请优先保持分类一致性。请忽略情感色彩，不要输出 comment。"
		}
	} else {
		if enableRoast {
			contextInstruction += "\n【重要指令】\n请务必在 'comment' 字段中填入一句简短、辛辣、幽默的吐槽（毒舌风格）。"
		} else {
			contextInstruction += "\n【重要指令】\n'comment' 字段是必填项，但请务必填入空字符串 \"\"，不要输出任何内容。"
		}
	}
	finalSystemPrompt := sysPrompt + contextInstruction

	req := openai.ChatCompletionRequest{
		Model: d.modelName,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: finalSystemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userContext},
		},
		// 注入动态工具
		Tools: []openai.Tool{
			GenerateBookExpenseTool(categories, enableRoast),
		},
		// 强制模型思考是否需要调用工具 (Auto 也是常用选项，Required 强制必须调)
		ToolChoice: openai.ToolChoice{
			Type: openai.ToolTypeFunction,
			Function: openai.ToolFunction{
				Name: "book_expense",
			},
		},
		Temperature: 0.1, // 低温有助于 JSON 格式稳定
		Stream:      true,
	}
	stream, err := d.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	outCh := make(chan string, 10)
	go func() {
		defer close(outCh)
		defer stream.Close()
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				slog.Error("Stream error", "err", err)
				return
			}
			if len(response.Choices) > 0 && len(response.Choices[0].Delta.ToolCalls) > 0 {
				fragment := response.Choices[0].Delta.ToolCalls[0].Function.Arguments
				if fragment != "" {
					outCh <- fragment
				}
			}
		}
	}()

	return outCh, nil
}
