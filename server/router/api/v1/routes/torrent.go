package routes

import (
	"github.com/gin-gonic/gin"

	"goto-bangumi/api/response"
)

// TorrentActionRequest 种子操作请求
type TorrentActionRequest struct {
	URL string `json:"url" binding:"required"`
}

// TorrentDownloadRequest 种子下载请求
type TorrentDownloadRequest struct {
	URL       string `json:"url" binding:"required"`
	BangumiID uint   `json:"bangumi_id,omitempty"`
	SavePath  string `json:"save_path,omitempty"`
}

// RegisterTorrentRoutes 注册种子管理路由
func RegisterTorrentRoutes(r *gin.RouterGroup) {
	torrent := r.Group("/torrent")
	{
		torrent.GET("/get_all", getAllTorrents)
		torrent.POST("/delete", deleteTorrent)
		torrent.POST("/disable", disableTorrent)
		torrent.POST("/download", downloadTorrent)
	}
}

// getAllTorrents 获取所有种子
// GET /api/v1/torrent/get_all?bangumi_id=xxx
func getAllTorrents(c *gin.Context) {
	// TODO: 实现获取所有种子逻辑
	response.Success(c, []any{})
}

// deleteTorrent 删除种子
// POST /api/v1/torrent/delete
func deleteTorrent(c *gin.Context) {
	var req TorrentActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", "无效的请求体")
		return
	}

	// TODO: 实现删除种子的逻辑
	response.SuccessWithMessage(c, "Torrent deleted successfully", "种子删除成功", nil)
}

// disableTorrent 禁用种子
// POST /api/v1/torrent/disable
func disableTorrent(c *gin.Context) {
	var req TorrentActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", "无效的请求体")
		return
	}

	// TODO: 实现禁用种子的逻辑
	response.SuccessWithMessage(c, "Torrent disabled successfully", "种子已禁用", nil)
}

// downloadTorrent 手动下载种子
// POST /api/v1/torrent/download
func downloadTorrent(c *gin.Context) {
	var req TorrentDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", "无效的请求体")
		return
	}

	// TODO: 实现手动下载种子的逻辑
	response.SuccessWithMessage(c, "Torrent download started", "开始下载种子", nil)
}
