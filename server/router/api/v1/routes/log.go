package routes

import (
	"bufio"
	"os"

	"github.com/gin-gonic/gin"

	"goto-bangumi/api/response"
)

const (
	// LogFilePath 日志文件路径
	LogFilePath = "./data/log/auto_bangumi.log"
	// MaxLogLines 最大返回行数
	MaxLogLines = 200
)

// RegisterLogRoutes 注册日志路由
func RegisterLogRoutes(r *gin.RouterGroup) {
	log := r.Group("/log")
	{
		log.GET("", getLog)
		log.GET("/clear", clearLog)
	}
}

// LogResponse 日志响应
type LogResponse struct {
	Lines []string `json:"lines"`
	Total int      `json:"total"`
}

// getLog 获取日志
// GET /api/v1/log
func getLog(c *gin.Context) {
	file, err := os.Open(LogFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			response.Success(c, LogResponse{
				Lines: []string{},
				Total: 0,
			})
			return
		}
		response.InternalError(c, "Failed to open log file", "无法打开日志文件")
		return
	}
	defer file.Close()

	// 读取所有行
	var allLines []string
	scanner := bufio.NewScanner(file)
	// 增加缓冲区大小以处理长行
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		response.InternalError(c, "Failed to read log file", "读取日志文件失败")
		return
	}

	// 返回最后 MaxLogLines 行
	total := len(allLines)
	start := 0
	if total > MaxLogLines {
		start = total - MaxLogLines
	}

	response.Success(c, LogResponse{
		Lines: allLines[start:],
		Total: total,
	})
}

// clearLog 清空日志
// GET /api/v1/log/clear
func clearLog(c *gin.Context) {
	// 截断文件
	file, err := os.OpenFile(LogFilePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			response.SuccessWithMessage(c, "Log file does not exist", "日志文件不存在", nil)
			return
		}
		response.InternalError(c, "Failed to clear log file", "清空日志文件失败")
		return
	}
	defer file.Close()

	response.SuccessWithMessage(c, "Log cleared successfully", "日志清空成功", nil)
}
