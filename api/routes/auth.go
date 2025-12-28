package routes

import (
	"github.com/gin-gonic/gin"

	"goto-bangumi/api/middleware"
	"goto-bangumi/api/response"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

// UserUpdateRequest 用户更新请求
type UserUpdateRequest struct {
	Username    string `json:"username,omitempty"`
	OldPassword string `json:"old_password,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
}

// TokenResponse Token 响应
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// RegisterAuthRoutes 注册认证路由
func RegisterAuthRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", login)
	}

	// 需要认证的路由
	authRequired := r.Group("/auth")
	authRequired.Use(middleware.JWTAuth())
	{
		authRequired.GET("/refresh_token", refreshToken)
		authRequired.GET("/logout", logout)
		authRequired.POST("/update", updateUser)
	}
}

// login 用户登录
// POST /api/v1/auth/login
func login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "Invalid request", "无效的请求参数")
		return
	}

	// TODO: 实现真正的用户验证逻辑
	token, _ := middleware.GenerateToken(req.Username)
	middleware.SetTokenCookie(c, token)

	response.Success(c, TokenResponse{
		AccessToken: token,
		TokenType:   "bearer",
	})
}

// refreshToken 刷新 Token
// GET /api/v1/auth/refresh_token
func refreshToken(c *gin.Context) {
	username := middleware.GetCurrentUser(c)
	token, _ := middleware.GenerateToken(username)
	middleware.SetTokenCookie(c, token)

	response.Success(c, TokenResponse{
		AccessToken: token,
		TokenType:   "bearer",
	})
}

// logout 用户登出
// GET /api/v1/auth/logout
func logout(c *gin.Context) {
	middleware.ClearTokenCookie(c)
	response.SuccessWithMessage(c, "Logged out successfully", "登出成功", nil)
}

// updateUser 更新用户信息
// POST /api/v1/auth/update
func updateUser(c *gin.Context) {
	// TODO: 实现用户更新逻辑
	response.SuccessWithMessage(c, "User updated successfully", "用户信息更新成功", nil)
}
