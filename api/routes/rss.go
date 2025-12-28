package routes

import (
	"github.com/gin-gonic/gin"

	"goto-bangumi/api/response"
)

// RSSAddRequest RSS 添加请求
type RSSAddRequest struct {
	URL       string `json:"url" binding:"required"`
	Name      string `json:"name,omitempty"`
	Aggregate bool   `json:"aggregate,omitempty"`
	Parser    string `json:"parser,omitempty"`
	Enabled   bool   `json:"enabled"`
	Filter    string `json:"filter,omitempty"`
	Include   string `json:"include,omitempty"`
}

// RSSUpdateRequest RSS 更新请求
type RSSUpdateRequest struct {
	URL       string `json:"url,omitempty"`
	Name      string `json:"name,omitempty"`
	Aggregate bool   `json:"aggregate,omitempty"`
	Parser    string `json:"parser,omitempty"`
	Enabled   *bool  `json:"enabled,omitempty"`
	Filter    string `json:"filter,omitempty"`
	Include   string `json:"include,omitempty"`
}

// RSSIDsRequest 批量操作请求
type RSSIDsRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// RSSAnalysisRequest RSS 分析请求
type RSSAnalysisRequest struct {
	URL       string `json:"url" binding:"required"`
	Aggregate bool   `json:"aggregate,omitempty"`
}

// RSSCollectRequest RSS 收集请求
type RSSCollectRequest struct {
	URL          string `json:"url" binding:"required"`
	OfficialName string `json:"official_name,omitempty"`
	Season       int    `json:"season,omitempty"`
}

// RSSSubscribeRequest RSS 订阅请求
type RSSSubscribeRequest struct {
	URL       string `json:"url" binding:"required"`
	BangumiID uint   `json:"bangumi_id,omitempty"`
	Season    int    `json:"season,omitempty"`
	Filter    string `json:"filter,omitempty"`
}

// RegisterRSSRoutes 注册 RSS 路由
func RegisterRSSRoutes(r *gin.RouterGroup) {
	rss := r.Group("/rss")
	{
		rss.GET("", getAllRSS)
		rss.POST("/add", addRSS)
		rss.POST("/enable/many", enableManyRSS)
		rss.DELETE("/delete/:id", deleteRSS)
		rss.POST("/delete/many", deleteManyRSS)
		rss.PATCH("/disable/:id", disableRSS)
		rss.POST("/disable/many", disableManyRSS)
		rss.PATCH("/update/:id", updateRSS)
		rss.GET("/refresh/all", refreshAllRSS)
		rss.GET("/torrent/:id", getRSSTorrents)
		rss.POST("/analysis", analysisRSS)
		rss.POST("/collect", collectRSS)
		rss.POST("/subscribe", subscribeRSS)
	}
}

// getAllRSS 获取所有 RSS 源
// GET /api/v1/rss
func getAllRSS(c *gin.Context) {
	// TODO: 实现获取所有 RSS 源逻辑
	response.Success(c, []any{})
}

// addRSS 添加 RSS 源
// POST /api/v1/rss/add
func addRSS(c *gin.Context) {
	var req RSSAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", "无效的请求体")
		return
	}

	// TODO: 实现添加 RSS 源逻辑
	response.SuccessWithMessage(c, "RSS added successfully", "RSS 添加成功", nil)
}

// enableManyRSS 批量启用 RSS
// POST /api/v1/rss/enable/many
func enableManyRSS(c *gin.Context) {
	var req RSSIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", "无效的请求体")
		return
	}

	// TODO: 实现批量启用 RSS 逻辑
	response.SuccessWithMessage(c, "RSS enabled successfully", "RSS 批量启用成功", nil)
}

// deleteRSS 删除 RSS
// DELETE /api/v1/rss/delete/:id
func deleteRSS(c *gin.Context) {
	// TODO: 实现删除 RSS 逻辑
	response.SuccessWithMessage(c, "RSS deleted successfully", "RSS 删除成功", nil)
}

// deleteManyRSS 批量删除 RSS
// POST /api/v1/rss/delete/many
func deleteManyRSS(c *gin.Context) {
	var req RSSIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", "无效的请求体")
		return
	}

	// TODO: 实现批量删除 RSS 逻辑
	response.SuccessWithMessage(c, "RSS deleted successfully", "RSS 批量删除成功", nil)
}

// disableRSS 禁用 RSS
// PATCH /api/v1/rss/disable/:id
func disableRSS(c *gin.Context) {
	// TODO: 实现禁用 RSS 逻辑
	response.SuccessWithMessage(c, "RSS disabled successfully", "RSS 已禁用", nil)
}

// disableManyRSS 批量禁用 RSS
// POST /api/v1/rss/disable/many
func disableManyRSS(c *gin.Context) {
	var req RSSIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", "无效的请求体")
		return
	}

	// TODO: 实现批量禁用 RSS 逻辑
	response.SuccessWithMessage(c, "RSS disabled successfully", "RSS 批量禁用成功", nil)
}

// updateRSS 更新 RSS
// PATCH /api/v1/rss/update/:id
func updateRSS(c *gin.Context) {
	var req RSSUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", "无效的请求体")
		return
	}

	// TODO: 实现更新 RSS 逻辑
	response.SuccessWithMessage(c, "RSS updated successfully", "RSS 更新成功", nil)
}

// refreshAllRSS 刷新所有 RSS
// GET /api/v1/rss/refresh/all
func refreshAllRSS(c *gin.Context) {
	// TODO: 实现刷新所有 RSS 的逻辑
	response.SuccessWithMessage(c, "RSS refresh started", "开始刷新 RSS", nil)
}

// getRSSTorrents 获取 RSS 源的种子列表
// GET /api/v1/rss/torrent/:id
func getRSSTorrents(c *gin.Context) {
	// TODO: 实现获取 RSS 关联的种子列表
	response.Success(c, []any{})
}

// analysisRSS 分析 RSS 源
// POST /api/v1/rss/analysis
func analysisRSS(c *gin.Context) {
	var req RSSAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", "无效的请求体")
		return
	}

	// TODO: 实现 RSS 分析逻辑
	result := struct {
		URL       string `json:"url"`
		Aggregate bool   `json:"aggregate"`
		Items     []any  `json:"items"`
	}{
		URL:       req.URL,
		Aggregate: req.Aggregate,
		Items:     []any{},
	}

	response.Success(c, result)
}

// collectRSS 收集番剧资源
// POST /api/v1/rss/collect
func collectRSS(c *gin.Context) {
	var req RSSCollectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", "无效的请求体")
		return
	}

	// TODO: 实现收集番剧所有资源的逻辑
	response.SuccessWithMessage(c, "Collection started", "开始收集资源", nil)
}

// subscribeRSS 订阅番剧
// POST /api/v1/rss/subscribe
func subscribeRSS(c *gin.Context) {
	var req RSSSubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", "无效的请求体")
		return
	}

	// TODO: 实现订阅番剧的逻辑
	response.SuccessWithMessage(c, "Subscription created", "订阅创建成功", nil)
}
