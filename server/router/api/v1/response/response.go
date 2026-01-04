package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	StatusCode int    `json:"status_code"`
	MsgEn      string `json:"msg_en"`
	MsgZh      string `json:"msg_zh"`
	Data       any    `json:"data,omitempty"`
}

// 预定义响应消息
var (
	MsgSuccess = Response{
		StatusCode: http.StatusOK,
		MsgEn:      "Success",
		MsgZh:      "成功",
	}

	MsgUnauthorized = Response{
		StatusCode: http.StatusUnauthorized,
		MsgEn:      "Unauthorized",
		MsgZh:      "未授权",
	}

	MsgBadRequest = Response{
		StatusCode: http.StatusBadRequest,
		MsgEn:      "Bad Request",
		MsgZh:      "请求错误",
	}

	MsgNotFound = Response{
		StatusCode: http.StatusNotFound,
		MsgEn:      "Not Found",
		MsgZh:      "未找到",
	}

	MsgInternalError = Response{
		StatusCode: http.StatusInternalServerError,
		MsgEn:      "Internal Server Error",
		MsgZh:      "服务器内部错误",
	}
)

// Success 返回成功响应
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		StatusCode: http.StatusOK,
		MsgEn:      "Success",
		MsgZh:      "成功",
		Data:       data,
	})
}

// SuccessWithMessage 返回带消息的成功响应
func SuccessWithMessage(c *gin.Context, msgEn, msgZh string, data any) {
	c.JSON(http.StatusOK, Response{
		StatusCode: http.StatusOK,
		MsgEn:      msgEn,
		MsgZh:      msgZh,
		Data:       data,
	})
}

// Error 返回错误响应
func Error(c *gin.Context, statusCode int, msgEn, msgZh string) {
	c.JSON(statusCode, Response{
		StatusCode: statusCode,
		MsgEn:      msgEn,
		MsgZh:      msgZh,
	})
}

// ErrorWithData 返回带数据的错误响应
func ErrorWithData(c *gin.Context, statusCode int, msgEn, msgZh string, data any) {
	c.JSON(statusCode, Response{
		StatusCode: statusCode,
		MsgEn:      msgEn,
		MsgZh:      msgZh,
		Data:       data,
	})
}

// BadRequest 返回 400 错误
func BadRequest(c *gin.Context, msgEn, msgZh string) {
	Error(c, http.StatusBadRequest, msgEn, msgZh)
}

// Unauthorized 返回 401 错误
func Unauthorized(c *gin.Context, msgEn, msgZh string) {
	Error(c, http.StatusUnauthorized, msgEn, msgZh)
}

// NotFound 返回 404 错误
func NotFound(c *gin.Context, msgEn, msgZh string) {
	Error(c, http.StatusNotFound, msgEn, msgZh)
}

// InternalError 返回 500 错误
func InternalError(c *gin.Context, msgEn, msgZh string) {
	Error(c, http.StatusInternalServerError, msgEn, msgZh)
}
