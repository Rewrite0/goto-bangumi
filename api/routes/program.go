package routes

import (
	"os"
	"runtime"

	"github.com/gin-gonic/gin"

	"goto-bangumi/api/response"
)

// Version 程序版本
const Version = "0.1.0"

// ProgramStatus 程序状态
type ProgramStatus struct {
	Running   bool   `json:"running"`
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// UpdateStatus 更新状态
type UpdateStatus struct {
	Updating bool   `json:"updating"`
	Progress int    `json:"progress"`
	Message  string `json:"message"`
}

// VersionInfo 版本信息
type VersionInfo struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	HasUpdate      bool   `json:"has_update"`
}

// DownloaderStatus 下载器状态
type DownloaderStatus struct {
	Connected bool   `json:"connected"`
	Type      string `json:"type"`
}

// RegisterProgramRoutes 注册程序控制路由
func RegisterProgramRoutes(r *gin.RouterGroup) {
	r.GET("/restart", restart)
	r.GET("/start", start)
	r.GET("/stop", stop)
	r.GET("/status", status)
	r.GET("/shutdown", shutdown)
	r.GET("/check/downloader", checkDownloader)
	r.GET("/check/update", checkUpdate)
	r.POST("/program/update", programUpdate)
	r.GET("/update/status", updateStatus)
}

// restart 重启程序
// GET /api/v1/restart
func restart(c *gin.Context) {
	// TODO: 实现重启逻辑
	response.SuccessWithMessage(c, "Restarting program", "正在重启程序", nil)
}

// start 启动程序
// GET /api/v1/start
func start(c *gin.Context) {
	// TODO: 实现启动逻辑
	response.SuccessWithMessage(c, "Program started", "程序已启动", nil)
}

// stop 停止程序
// GET /api/v1/stop
func stop(c *gin.Context) {
	// TODO: 实现停止逻辑
	response.SuccessWithMessage(c, "Program stopped", "程序已停止", nil)
}

// status 获取程序状态
// GET /api/v1/status
func status(c *gin.Context) {
	status := ProgramStatus{
		Running:   true,
		Version:   Version,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}

	response.Success(c, status)
}

// shutdown 关闭程序
// GET /api/v1/shutdown
func shutdown(c *gin.Context) {
	response.SuccessWithMessage(c, "Shutting down", "正在关闭程序", nil)

	go func() {
		os.Exit(0)
	}()
}

// checkDownloader 检查下载器状态
// GET /api/v1/check/downloader
func checkDownloader(c *gin.Context) {
	// TODO: 实现下载器检查逻辑
	status := DownloaderStatus{
		Connected: false,
		Type:      "unknown",
	}

	response.Success(c, status)
}

// checkUpdate 检查版本更新
// GET /api/v1/check/update
func checkUpdate(c *gin.Context) {
	// TODO: 实现版本检查逻辑
	info := VersionInfo{
		CurrentVersion: Version,
		LatestVersion:  Version,
		HasUpdate:      false,
	}

	response.Success(c, info)
}

// programUpdate 执行程序更新
// POST /api/v1/program/update
func programUpdate(c *gin.Context) {
	// TODO: 实现程序更新逻辑
	response.SuccessWithMessage(c, "Update started", "开始更新程序", nil)
}

// updateStatus 获取更新状态
// GET /api/v1/update/status
func updateStatus(c *gin.Context) {
	// TODO: 实现获取更新进度的逻辑
	status := UpdateStatus{
		Updating: false,
		Progress: 0,
		Message:  "No update in progress",
	}

	response.Success(c, status)
}
