// Package tasks 实现定时任务
package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"goto-bangumi/internal/conf"
	"goto-bangumi/internal/database"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/refresh"
)

// RSSRefreshTask RSS 刷新任务
type RSSRefreshTask struct {
	interval time.Duration
	enabled  bool
}

// NewRSSRefreshTask 创建 RSS 刷新任务
func NewRSSRefreshTask() *RSSRefreshTask {
	// 从配置中读取 RSS 刷新间隔
	programConfig := conf.GetConfigOrDefault("program", model.NewProgramConfig())
	interval := max(programConfig.RssTime, 900) // 最小间隔为 15 分钟

	task := &RSSRefreshTask{
		interval: time.Duration(interval) * time.Second,
		enabled:  true, // 默认启用
	}

	slog.Debug("创建 RSS 刷新任务", "间隔", task.interval)
	return task
}

// Name 返回任务名称
func (t *RSSRefreshTask) Name() string {
	return "RSS 刷新任务"
}

// Interval 返回执行间隔
func (t *RSSRefreshTask) Interval() time.Duration {
	return t.interval
}

// Enable 返回是否启用
func (t *RSSRefreshTask) Enable() bool {
	return t.enabled
}

// Run 执行 RSS 刷新
func (t *RSSRefreshTask) Run(ctx context.Context) error {
	db := database.GetDB()
	// 获取所有启用的 RSS 源
	rssList, err := db.ListActiveRSS()
	if err != nil {
		return fmt.Errorf("[RSS task] 获取 RSS 列表失败: %w", err)
	}

	slog.Debug("[Rss task] 开始刷新 RSS", "数量", len(rssList))

	// 刷新每个 RSS 源
	for _, rss := range rssList {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			slog.Debug("刷新 RSS 源", "名称", rss.Name, "URL", rss.URL)

			// 调用 refresh 模块的刷新方法
			refresh.RefreshRSS(ctx, rss.URL)

			// 为了避免短时间内请求过多，每个 RSS 源之间间隔一点时间
			time.Sleep(2 * time.Second)
		}
	}
	slog.Debug("RSS 刷新完成")
	return nil
}
