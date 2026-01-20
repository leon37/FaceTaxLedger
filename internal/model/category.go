package model

import (
	"strings"
)

// PredefinedCategories 预定义的分类列表，作为 AI 的参考
var PredefinedCategories = []string{
	"餐饮美食", "交通出行", "居家生活", "服饰美容",
	"休闲娱乐", "数码电器", "医疗健康", "人情往来",
	"学习教育", "金融保险", "其他消费",
}

// GetCategoryPrompt 生成 Prompt 用的分类提示词
func GetCategoryPrompt() string {
	return strings.Join(PredefinedCategories, ",")
}
