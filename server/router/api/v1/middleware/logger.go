package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if query != "" {
			path = path + "?" + query
		}

		slog.Info("HTTP 请求",
			"status", status,
			"method", c.Request.Method,
			"path", path,
			"latency", latency,
			"ip", c.ClientIP(),
		)
	}
}
