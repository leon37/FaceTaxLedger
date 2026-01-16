package main

func main() {
	// 1. 初始化依赖 (Infrastructure Layer)
	//apiKey := os.Getenv("DEEPSEEK_API_KEY")
	//// apiKey := "sk-..." // 如果环境变量搞不定，这里继续硬编码测试
	//llmClient := llm.NewDeepSeekClient(apiKey)
	//
	//// 2. 初始化服务 (Service Layer) -> 注入依赖
	//svc := service.NewExpenseService(llmClient)
	//
	//// 3. 模拟用户请求 (Controller Layer)
	//req := service.ExpenseInput{
	//	UserID:      "user_007",
	//	Amount:      1280.00,
	//	Description: "为了在相亲对象面前显摆，点了一瓶根本喝不懂的红酒",
	//}
	//
	//fmt.Println(">>> 正在提交账单给 Service 层...")
	//result, err := svc.SubmitExpense(context.Background(), req)
	//if err != nil {
	//	log.Fatalf("业务处理失败: %v", err)
	//}
	//
	//// 4. 打印最终结果 JSON
	//output, _ := json.MarshalIndent(result, "", "  ")
	//fmt.Println(string(output))
}
