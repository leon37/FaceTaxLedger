package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Authorization header required"})
			c.Abort()
			return
		}

		// 格式通常是 "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		secret := viper.GetString("jwt.secret")

		// 解析 Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 提取 Claims 并注入 Context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// JSON number parsing needs care, use float64 type assertion for default map claims
			if userIDStr, ok := claims["user_id"].(string); ok {
				c.Set("userID", userIDStr)
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Invalid token claims"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
