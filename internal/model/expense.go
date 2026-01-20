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
	UserID string  `gorm:"type:varchar(64);index" json:"user_id"`
	Amount float64 `gorm:"type:decimal(10,2)" json:"amount"`

	Comment  string `gorm:"type:text" json:"comment"`
	Category string `gorm:"type:varchar(64)" json:"category"`
	Note     string `gorm:"type:text" json:"note"`
}

// TableName 强制指定表名
func (ExpenseEntity) TableName() string {
	return "expenses"
}
