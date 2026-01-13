package model

// FaceTaxAnalysis 是 LLM 分析结果的结构化映射
// 这是一个核心领域模型，它决定了我们能把什么存进数据库
type FaceTaxAnalysis struct {
	// IsFaceTax: AI 判定的“面子税”布尔值
	// true 表示这笔消费是为了虚荣/人情/伪装
	IsFaceTax bool `json:"is_face_tax"`

	// TaxCategory: 面子税的具体分类
	// 仅当 IsFaceTax 为 true 时有值，例如 "酒精税", "人情税", "预支信心税"
	// 使用 omitempty，如果不是面子税，JSON 中该字段可能不存在
	TaxCategory string `json:"tax_category,omitempty"`

	// Comment: 毒舌评价
	// 这是产品的核心灵魂，必须展示给用户
	Comment string `json:"comment"`

	// SarcasmLevel: 嘲讽指数 (1-10)
	// 后期可以用于数据分析，比如"本月你承受了平均等级 8 的嘲讽"
	SarcasmLevel int `json:"sarcasm_level"`
}

// SystemPrompt 定义了 AI 的人设和输出协议
// 放在这里是为了让 Prompt 和 Struct 紧挨着，修改时能对照
const SystemPrompt = `你是一个尖酸刻薄、看透世俗、但真心为了主人好的老管家。
你的任务是根据用户的消费记录，分析其心理动机。

请严格遵守以下输出规则：
1. 严禁输出 JSON 格式以外的任何内容（不要 Markdown，不要解释）。
2. 输出必须符合以下 JSON 结构：
{
  "is_face_tax": true, 
  "tax_category": "酒精税", 
  "comment": "你的毒舌评价...", 
  "sarcasm_level": 8 
}`
