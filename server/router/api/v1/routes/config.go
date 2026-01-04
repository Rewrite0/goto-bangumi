package routes

import (
	"github.com/gin-gonic/gin"

	"goto-bangumi/api/response"
)

// RegisterConfigRoutes 注册配置路由
func RegisterConfigRoutes(r *gin.RouterGroup) {
	config := r.Group("/config")
	{
		config.GET("", getConfig)
		config.PUT("", updateConfig)
		config.POST("/test_notify", testNotify)
	}
}

// getConfig 获取配置
// GET /api/v1/config
func getConfig(c *gin.Context) {
	// TODO: 实现获取配置的逻辑
	response.Success(c, nil)
}

// updateConfig 更新配置
// PUT /api/v1/config
func updateConfig(c *gin.Context) {
	// TODO: 实现配置更新逻辑
	response.SuccessWithMessage(c, "Config updated successfully", "配置更新成功", nil)
}

// testNotify 测试通知
// POST /api/v1/config/test_notify
func testNotify(c *gin.Context) {
	// TODO: 实现测试通知逻辑
	response.SuccessWithMessage(c, "Test notification sent successfully", "测试通知发送成功", nil)
}
