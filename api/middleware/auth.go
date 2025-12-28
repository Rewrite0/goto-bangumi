package middleware

import (
	"github.com/gin-gonic/gin"
)

const (
	// TokenContextKey Token 在 context 中的 key
	TokenContextKey = "user"
)

// JWTAuth JWT 认证中间件
// TODO: 实现真正的 JWT 认证逻辑
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 暂时跳过认证，直接通过
		c.Set(TokenContextKey, "admin")
		c.Next()
	}
}

// GetCurrentUser 获取当前用户
func GetCurrentUser(c *gin.Context) string {
	if user, exists := c.Get(TokenContextKey); exists {
		return user.(string)
	}
	return ""
}

// GenerateToken 生成 Token
// TODO: 实现真正的 JWT Token 生成
func GenerateToken(username string) (string, error) {
	return "mock_token_" + username, nil
}

// SetTokenCookie 设置 Token Cookie
func SetTokenCookie(c *gin.Context, token string) {
	c.SetCookie(
		"access_token",
		token,
		86400, // 1 day
		"/",
		"",
		false,
		true,
	)
}

// ClearTokenCookie 清除 Token Cookie
func ClearTokenCookie(c *gin.Context) {
	c.SetCookie(
		"access_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)
}
