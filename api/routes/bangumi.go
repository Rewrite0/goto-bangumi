package routes

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"goto-bangumi/api/response"
)

const (
	// PosterBasePath 海报基础路径
	PosterBasePath = "./data/posters"
)

// BangumiUpdateRequest 番剧更新请求
type BangumiUpdateRequest struct {
	OfficialTitle string `json:"official_title,omitempty"`
	TitleRaw      string `json:"title_raw,omitempty"`
	Season        int    `json:"season,omitempty"`
	SeasonRaw     string `json:"season_raw,omitempty"`
	Group         string `json:"group,omitempty"`
	Offset        int    `json:"offset,omitempty"`
	Filter        string `json:"filter,omitempty"`
	RSSLink       string `json:"rss_link,omitempty"`
	PosterLink    string `json:"poster_link,omitempty"`
	Enabled       *bool  `json:"enabled,omitempty"`
	SavePath      string `json:"save_path,omitempty"`
}

// BangumiIDsRequest 批量操作请求
type BangumiIDsRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// RegisterBangumiRoutes 注册番剧管理路由
func RegisterBangumiRoutes(r *gin.RouterGroup) {
	bangumi := r.Group("/bangumi")
	{
		bangumi.GET("/get/all", getAllBangumi)
		bangumi.GET("/get/:id", getBangumi)
		bangumi.PATCH("/update/:id", updateBangumi)
		bangumi.DELETE("/delete/:id", deleteBangumi)
		bangumi.DELETE("/delete/many", deleteManyBangumi)
		bangumi.DELETE("/disable/:id", disableBangumi)
		bangumi.DELETE("/disable/many", disableManyBangumi)
		bangumi.GET("/enable/:id", enableBangumi)
		bangumi.GET("/refresh/poster/all", refreshAllPosters)
		bangumi.GET("/reset/all", resetAllBangumi)
		bangumi.GET("/posters/*path", getPoster)
	}
}

// getAllBangumi 获取所有番剧
// GET /api/v1/bangumi/get/all
func getAllBangumi(c *gin.Context) {
	// TODO: 实现获取所有番剧逻辑
	response.Success(c, []any{})
}

// getBangumi 获取指定番剧
// GET /api/v1/bangumi/get/:id
func getBangumi(c *gin.Context) {
	// TODO: 实现获取指定番剧逻辑
	response.Success(c, nil)
}

// updateBangumi 更新番剧规则
// PATCH /api/v1/bangumi/update/:id
func updateBangumi(c *gin.Context) {
	var req BangumiUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", "无效的请求体")
		return
	}

	// TODO: 实现更新番剧逻辑
	response.SuccessWithMessage(c, "Bangumi updated successfully", "番剧更新成功", nil)
}

// deleteBangumi 删除番剧
// DELETE /api/v1/bangumi/delete/:id
func deleteBangumi(c *gin.Context) {
	// TODO: 实现删除番剧逻辑
	response.SuccessWithMessage(c, "Bangumi deleted successfully", "番剧删除成功", nil)
}

// deleteManyBangumi 批量删除番剧
// DELETE /api/v1/bangumi/delete/many
func deleteManyBangumi(c *gin.Context) {
	var req BangumiIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", "无效的请求体")
		return
	}

	// TODO: 实现批量删除番剧逻辑
	response.SuccessWithMessage(c, "Bangumi deleted successfully", "番剧批量删除成功", nil)
}

// disableBangumi 禁用番剧
// DELETE /api/v1/bangumi/disable/:id
func disableBangumi(c *gin.Context) {
	// TODO: 实现禁用番剧逻辑
	response.SuccessWithMessage(c, "Bangumi disabled successfully", "番剧已禁用", nil)
}

// disableManyBangumi 批量禁用番剧
// DELETE /api/v1/bangumi/disable/many
func disableManyBangumi(c *gin.Context) {
	var req BangumiIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", "无效的请求体")
		return
	}

	// TODO: 实现批量禁用番剧逻辑
	response.SuccessWithMessage(c, "Bangumi disabled successfully", "番剧批量禁用成功", nil)
}

// enableBangumi 启用番剧
// GET /api/v1/bangumi/enable/:id
func enableBangumi(c *gin.Context) {
	// TODO: 实现启用番剧逻辑
	response.SuccessWithMessage(c, "Bangumi enabled successfully", "番剧已启用", nil)
}

// refreshAllPosters 刷新所有海报
// GET /api/v1/bangumi/refresh/poster/all
func refreshAllPosters(c *gin.Context) {
	// TODO: 实现刷新所有海报的逻辑
	response.SuccessWithMessage(c, "Poster refresh started", "开始刷新海报", nil)
}

// resetAllBangumi 重置所有番剧规则
// GET /api/v1/bangumi/reset/all
func resetAllBangumi(c *gin.Context) {
	// TODO: 实现重置所有番剧规则的逻辑
	response.SuccessWithMessage(c, "All bangumi rules reset", "所有番剧规则已重置", nil)
}

// getPoster 获取海报图片
// GET /api/v1/bangumi/posters/*path
func getPoster(c *gin.Context) {
	path := c.Param("path")

	// 安全检查：防止目录遍历
	if strings.Contains(path, "..") {
		response.BadRequest(c, "Invalid path", "无效的路径")
		return
	}

	// 构建完整路径
	fullPath := filepath.Join(PosterBasePath, path)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		response.NotFound(c, "Poster not found", "海报未找到")
		return
	}

	// 返回文件
	c.File(fullPath)
}
