package api

import (
	"github.com/gin-gonic/gin"
	"github.com/leon37/FaceTaxLedger/internal/api/controller"
	"github.com/leon37/FaceTaxLedger/internal/api/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/leon37/FaceTaxLedger/docs"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine, authCtrl *controller.AuthController, expenseCtrl *controller.ExpenseController) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	public := r.Group("/api/v1/auth")
	{
		public.POST("/register", authCtrl.Register)
		public.POST("/login", authCtrl.Login)
	}

	// API 组
	protected := r.Group("/api/v1")
	protected.Use(middleware.JWTAuth())
	{
		protected.POST("/expenses/analyze", expenseCtrl.Analyze)
		protected.GET("/expenses", expenseCtrl.List)
		protected.POST("/expenses/delete", expenseCtrl.Delete)
		protected.POST("/expenses/update", expenseCtrl.Update)
	}
}
