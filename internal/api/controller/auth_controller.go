package controller

import (
	"github.com/leon37/FaceTaxLedger/internal/api/response"
	"github.com/leon37/FaceTaxLedger/internal/service"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthController 处理用户认证
type AuthController struct {
	authService *service.AuthService
}

// NewAuthController 构造函数
func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// ==========================================
// DTOs (请求/响应参数定义)
// ==========================================

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

// ==========================================
// Handlers
// ==========================================

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户，密码加密存储
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册参数"
// @Success 200 {object} response.Response "Code=0 成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /auth/register [post]
func (ctrl *AuthController) Register(c *gin.Context) {
	var req RegisterRequest

	// 1. 参数校验
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("Register params invalid", "err", err)
		// 参数错误返回 400
		response.Error(c, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	// 2. 业务逻辑
	err := ctrl.authService.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		slog.Error("Register failed", "email", req.Email, "err", err)

		// 简单的错误处理：如果有特定的重复错误，可以判断 err 内容
		// 这里统一返回 200 + 错误提示，或根据需求返回 409 Conflict
		response.Error(c, http.StatusOK, "注册失败: "+err.Error())
		return
	}

	// 3. 成功响应
	slog.Info("User registered", "email", req.Email)
	response.Success(c, nil) // data 为 nil，只返回 code:0 msg:success
}

// Login 用户登录
// @Summary 用户登录
// @Description 校验账号密码，颁发 JWT Token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录参数"
// @Success 200 {object} response.Response{data=LoginResponse} "包含 Token 和 UserID"
// @Router /auth/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	var req LoginRequest

	// 1. 参数校验
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数格式错误")
		return
	}

	// 2. 业务逻辑
	token, userID, err := ctrl.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		slog.Warn("Login failed", "email", req.Email, "err", err)
		// 登录失败通常返回 200(业务错误) 或 401(未授权)，取决于前端约定
		// 为了防止暴力破解，提示信息模糊化
		response.Error(c, http.StatusOK, "登录失败: 账号或密码错误")
		return
	}

	// 3. 成功响应
	slog.Info("User logged in", "userID", userID)
	response.Success(c, LoginResponse{
		Token:  token,
		UserID: userID,
	})
}
