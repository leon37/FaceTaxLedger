package llm

import (
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// GenerateBookExpenseTool 动态生成记账工具定义
// categories: 包含预定义分类和用户自定义分类
func GenerateBookExpenseTool(categories []string, enableRoast bool) openai.Tool {
	commentDesc := "毒舌评价。"
	if enableRoast {
		commentDesc = "基于消费内容的一句简短、辛辣、幽默的吐槽。"
	} else {
		commentDesc = "必须严格返回空字符串 \"\"，禁止包含任何字符。"
	}

	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "book_expense",
			Description: "记录用户的单笔消费详情，提取金额、日期、分类和备注。",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"amount": {
						Type:        jsonschema.Number,
						Description: "消费的总金额，如果是多笔消费请自动求和。",
					},
					"category": {
						Type:        jsonschema.String,
						Enum:        categories, // 核心：动态注入 Enum
						Description: "消费类别，必须严格匹配列表中的一项。",
					},
					"date": {
						Type:        jsonschema.String,
						Description: "消费发生的日期 (YYYY-MM-DD)。基于当前时间推断（如'昨天'）。",
					},
					"note": {
						Type:        jsonschema.String,
						Description: "消费内容的简短纯粹描述，去除金额和时间词。",
					},
					"comment": {
						Type:        jsonschema.String,
						Description: commentDesc,
					},
				},
				// 强制模型必须返回这些字段
				Required: []string{"amount", "category", "date", "note", "comment"},
			},
		},
	}
}
