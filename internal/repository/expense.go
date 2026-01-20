package repository

import (
	"context"
	"time"

	"github.com/leon37/FaceTaxLedger/internal/model"
	"gorm.io/gorm"
)

// ExpenseRepo 定义接口 (为了以后方便 Mock)
type ExpenseRepo interface {
	Create(ctx context.Context, expense *model.ExpenseEntity) error
	List(ctx context.Context, filter ExpenseFilter) ([]model.ExpenseEntity, int64, error)
	GetByID(ctx context.Context, id int64) (*model.ExpenseEntity, error)
	Update(ctx context.Context, expense *model.ExpenseEntity) error
	Delete(ctx context.Context, id int64) error
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

func (r *expenseRepo) List(ctx context.Context, filter ExpenseFilter) ([]model.ExpenseEntity, int64, error) {
	var expenses []model.ExpenseEntity
	var total int64

	// 1. 构建基础查询 (带上 Context 和 UserID)
	db := r.db.WithContext(ctx).Model(&model.ExpenseEntity{}).Where("user_id = ?", filter.UserID)

	// 2. 动态追加条件
	if filter.Category != "" {
		db = db.Where("category = ?", filter.Category)
	}
	if !filter.StartDate.IsZero() {
		db = db.Where("created_at >= ?", filter.StartDate)
	}
	if !filter.EndDate.IsZero() {
		db = db.Where("created_at <= ?", filter.EndDate)
	}

	// 3. 计算总数 (在分页之前)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 4. 分页与排序 (按时间倒序)
	offset := (filter.Page - 1) * filter.PageSize
	err := db.Order("created_at DESC").
		Limit(filter.PageSize).
		Offset(offset).
		Find(&expenses).Error

	return expenses, total, err
}

func (r *expenseRepo) GetByID(ctx context.Context, id int64) (*model.ExpenseEntity, error) {
	var expense model.ExpenseEntity
	err := r.db.WithContext(ctx).First(&expense, id).Error
	return &expense, err
}

func (r *expenseRepo) Update(ctx context.Context, expense *model.ExpenseEntity) error {
	return r.db.WithContext(ctx).Save(expense).Error
}

func (r *expenseRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.ExpenseEntity{}, id).Error
}

type ExpenseFilter struct {
	UserID    string
	Category  string    // 可选
	StartDate time.Time // 可选
	EndDate   time.Time // 可选
	Page      int       // 分页：第几页
	PageSize  int       // 分页：每页多少条
}
