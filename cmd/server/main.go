package main

import (
	"context"
	"github.com/leon37/FaceTaxLedger/internal/api"
	"github.com/leon37/FaceTaxLedger/internal/api/controller"
	"github.com/leon37/FaceTaxLedger/internal/infrastructure/embedding"
	"github.com/leon37/FaceTaxLedger/internal/infrastructure/vectordb"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leon37/FaceTaxLedger/internal/config"
	"github.com/leon37/FaceTaxLedger/internal/infrastructure/database"
	"github.com/leon37/FaceTaxLedger/internal/infrastructure/llm"
	"github.com/leon37/FaceTaxLedger/internal/repository"
	"github.com/leon37/FaceTaxLedger/internal/service"
)

// @title           FaceTax API
// @version         1.0
// @description     基于 Go + Gin + Qdrant 的 AI 智能记账系统
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 请在输入框中输入 "Bearer <token>" (注意 Bearer 和 token 之间有空格)

func main() {
	// 1. 初始化 Logger
	// 使用 JSONHandler 可以让日志以 JSON 格式输出，方便解析
	// AddSource: true 会在日志里显示文件名和行号
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug, // 开发阶段设为 Debug，生产环境改为 Info
	}))

	// 设置为全局默认 logger
	slog.SetDefault(logger)

	slog.Info("FaceTax 系统启动中...")

	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}
	log.Println("配置加载成功")

	// 2. Infra Initialization
	llmClient := llm.NewDeepSeekClient(conf.DeepSeek.APIKey, conf.DeepSeek.BaseURL, conf.DeepSeek.Model)
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

	if conf.Server.Port != ":8080" { // 简单的判断，生产环境建议用配置字段
		gin.SetMode(gin.ReleaseMode)
	}

	embedder := embedding.NewOpenAIClient(conf.OpenAI.APIKey, conf.OpenAI.BaseURL, conf.OpenAI.Model)
	// 3. Layer Wiring (依赖注入)
	repo := repository.NewExpenseRepo(db)
	memoryRepo := vectordb.NewQdrantRepository(vecClient)
	svc := service.NewExpenseService(llmClient, embedder, repo, memoryRepo) // 注入 repo

	// 4. Server Start
	r := gin.Default()
	expenseController := controller.NewExpenseController(svc)

	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo)
	authController := controller.NewAuthController(authSvc)
	api.RegisterRoutes(r, authController, expenseController)

	slog.Info("FaceTax Web Server 启动中", "port", conf.Server.Port)
	if err := r.Run(conf.Server.Port); err != nil {
		slog.Error("服务器启动失败", "error", err)
	}
}
