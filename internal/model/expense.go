package model

import (
	"time"

	"gorm.io/gorm"
)

// ExpenseEntity 是映射数据库表的结构体
type ExpenseEntity struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 输入数据
	UserID      string  `gorm:"type:varchar(64);index" json:"user_id"`
	Amount      float64 `gorm:"type:decimal(10,2)" json:"amount"`
	Description string  `gorm:"type:text" json:"description"`

	// AI 分析结果 (把 FaceTaxAnalysis 拍平存进去)
	IsFaceTax    bool   `json:"is_face_tax"`
	TaxCategory  string `gorm:"type:varchar(64)" json:"tax_category"`
	Comment      string `gorm:"type:text" json:"comment"`
	SarcasmLevel int    `json:"sarcasm_level"`
}

// TableName 强制指定表名
func (ExpenseEntity) TableName() string {
	return "expenses"
}
