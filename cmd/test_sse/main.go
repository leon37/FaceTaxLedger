package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func main() {
	url := "http://localhost:8080/api/v1/expenses/analyze" // 替换你的真实路由
	payload := map[string]string{
		"description": "打车花了50元，太堵了", // 测试输入
	}
	jsonData, _ := json.Marshal(payload)

	// 1. 发起 POST 请求
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream") // 告诉服务器我要流
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjkxNjczODgsInVzZXJfaWQiOiIwMTliZGIwNi0xZjU4LTc2NzEtOGUwMC1lN2RlNDBlNzY2NzAifQ.nYh7dhO449lt8ww8dAusj3tRjkYQdOm2bjnsaS0VhmI")
	// 如果有鉴权，记得加 Header: req.Header.Set("Authorization", "Bearer ...")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("✅ 连接建立，开始接收流...")
	fmt.Println("--------------------------------")

	// 2. 使用 Scanner 按行读取 (SSE 是按行传输的)
	scanner := bufio.NewScanner(resp.Body)

	var fullBuffer strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// 忽略空行
		if line == "" {
			continue
		}

		fmt.Printf("[收到原始数据] %s\n", line)

		// 3. 解析 SSE 协议 (格式通常是 "event: xxx" 或 "data: xxx")
		if strings.HasPrefix(line, "event: delta") {
			// 下一行通常是 data
			continue
		}

		if strings.HasPrefix(line, "data:") {
			content := strings.TrimPrefix(line, "data: ")
			fmt.Printf("   └──> 解析内容: %s\n", content)
			fullBuffer.WriteString(content)
		}

		if strings.HasPrefix(line, "event: done") {
			fmt.Println("\n🏁 流传输结束 (Done Signal)")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("读取流错误:", err)
	}

	fmt.Println("--------------------------------")
	fmt.Println("📝 最终拼接结果:", fullBuffer.String())
}
