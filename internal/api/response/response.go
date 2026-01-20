package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code int         `json:"code"` // 0 代表成功，非 0 代表错误码
	Msg  string      `json:"msg"`  // 提示信息
	Data interface{} `json:"data"` // 数据载荷
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

// Error 错误响应
func Error(c *gin.Context, httpStatus int, msg string) {
	c.JSON(httpStatus, Response{
		Code: -1, // 这里可以根据业务定义具体的错误码
		Msg:  msg,
		Data: nil,
	})
}
