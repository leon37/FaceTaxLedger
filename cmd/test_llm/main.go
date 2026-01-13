package main

import (
	"context"
	"fmt"
	"github.com/leon37/FaceTaxLedger/internal/infrastructure/llm"
	"log"
	"os"
)

func main() {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set DEEPSEEK_API_KEY env")
	}

	client := llm.NewDeepSeekClient(apiKey)

	// 模拟场景
	input := "昨晚为了在老同学面前显摆，抢着买了单，花了800块，现在想吃土。"

	fmt.Println("正在让管家分析你的面子税...")
	result, err := client.AnalyzeExpense(context.Background(), input)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	fmt.Printf("=== 毒舌账单 ===\n")
	fmt.Printf("是否面子税: %v\n", result.IsFaceTax)
	fmt.Printf("分类: %s\n", result.TaxCategory)
	fmt.Printf("管家评价: %s\n", result.Comment)
	fmt.Printf("嘲讽指数: %d/10\n", result.SarcasmLevel)
}
