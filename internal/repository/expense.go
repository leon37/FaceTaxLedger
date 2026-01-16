package repository

import (
	"context"

	"github.com/leon37/FaceTaxLedger/internal/model"
	"gorm.io/gorm"
)

// ExpenseRepo 定义接口 (为了以后方便 Mock)
type ExpenseRepo interface {
	Create(ctx context.Context, expense *model.ExpenseEntity) error
}

// expenseRepo 实现
type expenseRepo struct {
	db *gorm.DB
}

// NewExpenseRepo 构造函数
func NewExpenseRepo(db *gorm.DB) ExpenseRepo {
	return &expenseRepo{db: db}
}

// Create 插入一条记录
func (r *expenseRepo) Create(ctx context.Context, expense *model.ExpenseEntity) error {
	// WithContext 确保请求超时能传递到数据库层
	return r.db.WithContext(ctx).Create(expense).Error
}
