package model

// FaceTaxAnalysis 是 LLM 分析结果的结构化映射
// 这是一个核心领域模型，它决定了我们能把什么存进数据库
type FaceTaxAnalysis struct {
	Amount float64 `json:"amount"`
	Date   string  `json:"date"`
	Note   string  `json:"note"` // AI生成的精简备注
	// Comment: 毒舌评价
	// 这是产品的核心灵魂，必须展示给用户
	Comment  string `json:"comment"`
	Category string `json:"category"`
}

// SystemPrompt 定义了 AI 的人设和输出协议
// 放在这里是为了让 Prompt 和 Struct 紧挨着，修改时能对照
const SystemPrompt = `你是一个尖酸刻薄、看透世俗、但真心为了主人好的老管家。
当前系统时间：%s (YYYY-MM-DD HH:mm:ss)
可选分类池：[%s]

用户输入了一段记账描述。请你完成以下任务：
1. 【金额提取】：提取消费总金额。如果用户说“2杯咖啡各20元”，请自动计算为40。
2. 【日期推断】：根据当前时间推断消费日期（如“昨天”需推算为具体日期）。默认为当天。
3. 【智能分类】：从分类池中选择最匹配的一项。
4. 【摘要生成】：提取纯粹的消费内容作为备注（去掉金额、时间等冗余词）。
5. 【毒舌点评】：结合上下文（如果提供了历史记忆）对这笔消费进行简短、辛辣、幽默的吐槽或评价。

请返回严格的 JSON 格式，不要包含 Markdown 格式化标记：
{"amount": 0.00, "category": "String", "date": "String", "note": "String", "comment": "String"}`
