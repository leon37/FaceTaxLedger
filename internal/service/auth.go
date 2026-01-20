package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/leon37/FaceTaxLedger/internal/model"
	"github.com/leon37/FaceTaxLedger/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// Register 注册逻辑
func (s *AuthService) Register(ctx context.Context, username, email, password string) error {
	// 1. 检查是否存在 (略，DB Unique Index 会兜底，最好先查一下)

	// 2. 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 3. 落库
	id, _ := uuid.NewV7()
	user := &model.User{
		ID:       id.String(),
		Username: username,
		Email:    email,
		Password: string(hash),
	}
	return s.userRepo.Create(ctx, user)
}

// Login 登录逻辑，返回 Token
func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	// 1. 查用户
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials") // 模糊报错为了安全
	}

	// 2. 比对密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	// 3. 生成 JWT
	return s.generateToken(user.ID)
}

func (s *AuthService) generateToken(userID string) (string, string, error) {
	secret := viper.GetString("jwt.secret")
	expireHours := viper.GetInt("jwt.expire_hours")

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * time.Duration(expireHours)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(secret))
	return ss, userID, err
}
