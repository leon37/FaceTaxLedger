package main

import (
	"context"
	"fmt"
	"github.com/leon37/FaceTaxLedger/internal/config"
	"log"
	"os"
	"time"

	"github.com/leon37/FaceTaxLedger/internal/infrastructure/llm"
	// 记得替换 import 路径为你的实际项目路径
)

func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}
	log.Println("配置加载成功")
	// 1. 初始化 Client (从环境变量读取 Key，安全第一)
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置环境变量 DEEPSEEK_API_KEY")
	}
	llmClient := llm.NewDeepSeekClient(conf.DeepSeek.APIKey, conf.DeepSeek.BaseURL, conf.DeepSeek.Model)

	// 2. 准备测试数据
	ctx := context.Background()

	// 模拟：预定义分类 + 用户自定义的一个生僻分类
	categories := []string{"餐饮", "交通", "购物", "数码产品", "氪金手游"}

	// 模拟：RAG 检索到的历史数据
	historyLogs := []string{
		"2026-01-20 (3天前) [氪金手游] 原神充值: 又是小保底歪了",
	}

	// 3. 定义测试用例
	testCases := []struct {
		Name        string
		Input       string
		EnableRoast bool
	}{
		{
			Name:        "场景1：普通记账+毒舌",
			Input:       "刚才打车回家花了25块钱，司机开太快了",
			EnableRoast: true,
		},
		{
			Name:        "场景2：自定义分类+RAG参考",
			Input:       "冲了个月卡30元",
			EnableRoast: true,
		},
		{
			Name:        "场景3：多笔消费自动求和+关闭毒舌",
			Input:       "买了2杯拿铁，一杯20",
			EnableRoast: false,
		},
	}

	// 4. 执行循环测试
	for _, tc := range testCases {
		fmt.Printf("\n-------- 测试: %s --------\n", tc.Name)
		fmt.Printf("输入: %s\n", tc.Input)

		start := time.Now()
		result, err := llmClient.AnalyzeExpense(ctx, tc.Input, categories, historyLogs, tc.EnableRoast)
		duration := time.Since(start)

		if err != nil {
			log.Printf("❌ 调用失败: %v\n", err)
			continue
		}

		fmt.Printf("✅ 调用成功 (耗时 %v)\n", duration)
		fmt.Printf("提取金额: %.2f\n", result.Amount)
		fmt.Printf("提取分类: %s\n", result.Category)
		fmt.Printf("提取日期: %s\n", result.Date)
		fmt.Printf("提取摘要: %s\n", result.Note)
		fmt.Printf("毒舌吐槽: %s\n", result.Comment)
	}
}
