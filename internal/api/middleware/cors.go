package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")

		// 1. 设置允许的 Origin
		// 生产环境建议将 "*" 替换为具体的域名，开发环境如果 Origin 为 null (本地文件打开)，也可以暂时设为 "*"
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		// 2. 设置其他允许的 Header
		c.Header("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		// 3. 【关键修复】如果是 OPTIONS 请求，直接终止并返回 204
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return // 必须 return，不再往下执行
		}

		// 4. 非 OPTIONS 请求，继续处理业务
		c.Next()
	}
}
