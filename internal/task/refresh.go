// Package tasks 实现定时任务
package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/refresh"
	"goto-bangumi/internal/taskrunner"
)

// RSSRefreshTask RSS 刷新任务
type RSSRefreshTask struct {
	interval  time.Duration
	enabled   bool
	runner    *taskrunner.TaskRunner
	db        *database.DB
	refresher *refresh.Refresher
}

// NewRSSRefreshTask 创建 RSS 刷新任务
func NewRSSRefreshTask(programConfig model.ProgramConfig, runner *taskrunner.TaskRunner, db *database.DB, refresher *refresh.Refresher) *RSSRefreshTask {
	interval := programConfig.RssTime

	task := &RSSRefreshTask{
		interval:  time.Duration(interval) * time.Second,
		enabled:   true,
		runner:    runner,
		db:        db,
		refresher: refresher,
	}

	slog.Debug("[task rss]创建 RSS 刷新任务", "间隔", task.interval)
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
	// 获取所有启用的 RSS 源
	rssList, err := t.db.ListActiveRSS(ctx)
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
			slog.Debug("[refresh] 刷新 RSS 源", "名称", rss.Name, "URL", rss.Link)
			t.refresher.FindNewBangumi(ctx, rss)

			// 调用 refresh 模块的刷新方法
			t.refresher.RefreshRSS(ctx, rss.Link, t.runner)

			// 为了避免短时间内请求过多，每个 RSS 源之间间隔一点时间
			time.Sleep(2 * time.Second)
		}
	}
	slog.Debug("RSS 刷新完成")
	return nil
}
