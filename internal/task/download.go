// 一些定时的任务
package task

import (
	"context"
	"log/slog"
	"time"

	"goto-bangumi/internal/download"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/taskrunner"
)

// DownloadTask 下载任务
type DownloadTask struct {
	interval time.Duration
	enabled  bool
	runner   *taskrunner.TaskRunner
}

// NewDownloadTask 创建下载任务
func NewDownloadTask(runner *taskrunner.TaskRunner) *DownloadTask {
	return &DownloadTask{
		interval: 5 * time.Second,
		enabled:  true,
		runner:   runner,
	}
}

// Name 返回任务名称
func (t *DownloadTask) Name() string {
	return "下载任务"
}

// Interval 返回执行间隔
func (t *DownloadTask) Interval() time.Duration {
	return t.interval
}

// Enable 返回是否启用
func (t *DownloadTask) Enable() bool {
	return t.enabled
}

// Run 从队列取出种子，创建任务提交到 taskrunner
func (t *DownloadTask) Run(ctx context.Context) error {
	select {
	case tb := <-download.DQueue.Queue:
		torrent := tb.Torrent
		bangumi := tb.Bangumi
		download.DQueue.InQueue.Delete(torrent.Link)

		slog.Debug("[download task] 提交任务到 taskrunner", "Name", torrent.Name)
		task := model.NewTask(torrent, bangumi)
		t.runner.Submit(task)
	default:
	}
	return nil
}
