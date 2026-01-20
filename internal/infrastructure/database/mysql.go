package database

import (
	"log"
	"time"

	"github.com/leon37/FaceTaxLedger/internal/model" // 替换为你的 module 名
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewMySQLConnection(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 开发阶段显示 SQL 日志
	})
	if err != nil {
		log.Fatalf("Fatal: 无法连接数据库: %v", err)
	}

	// 自动建表 (Auto Migrate)
	// 这是 GORM 最爽的功能，自动在 MySQL 里创建 expenses 表
	if err := db.AutoMigrate(&model.ExpenseEntity{}); err != nil {
		log.Fatalf("Fatal: 数据库迁移失败: %v", err)
	}

	if err = db.AutoMigrate(&model.User{}); err != nil {
		log.Fatalf("Fatal: 数据库迁移失败: %v", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db
}
