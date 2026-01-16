package main

import (
	"context"
	"github.com/leon37/FaceTaxLedger/internal/infrastructure/embedding"
	"github.com/leon37/FaceTaxLedger/internal/infrastructure/vectordb"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leon37/FaceTaxLedger/internal/config"
	"github.com/leon37/FaceTaxLedger/internal/handler"
	"github.com/leon37/FaceTaxLedger/internal/infrastructure/database"
	"github.com/leon37/FaceTaxLedger/internal/infrastructure/llm"
	"github.com/leon37/FaceTaxLedger/internal/repository"
	"github.com/leon37/FaceTaxLedger/internal/service"
)

func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}
	log.Println("配置加载成功")

	// 2. Infra Initialization
	llmClient := llm.NewDeepSeekClient(conf.DeepSeek.APIKey)
	db := database.NewMySQLConnection(conf.Database.DSN) // 这里会自动建表

	vecClient, err := vectordb.NewQdrantClient(conf.Qdrant.Host, conf.Qdrant.Port)
	if err != nil {
		log.Fatalf("Failed to init Vector DB: %v", err)
	}
	defer vecClient.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := vecClient.InitCollection(ctx); err != nil {
		// 如果初始化失败（比如连不上，或者无法创建），直接崩盘退出
		// 这是为了防止后续业务运行时报错
		log.Fatalf("Failed to init Qdrant collection: %v", err)
	}

	embedder := embedding.NewOpenAIClient(conf.OpenAI.APIKey, conf.OpenAI.BaseURL, conf.OpenAI.Model)
	// 3. Layer Wiring (依赖注入)
	repo := repository.NewExpenseRepo(db)
	memoryRepo := vectordb.NewQdrantRepository(vecClient, embedder)
	svc := service.NewExpenseService(llmClient, repo, memoryRepo) // 注入 repo
	h := handler.NewExpenseHandler(svc)

	// 4. Server Start
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.POST("/expenses", h.HandleSubmitExpense)
	}

	log.Println("FaceTax Server running...")
	r.Run(":8080")
}
