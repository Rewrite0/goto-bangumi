// 一些定时的任务
package task


import (
	"context"
	"log/slog"
	"path/filepath"
	"strconv"
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

// Run 执行下载任务
func (t *DownloadTask) Run(ctx context.Context) error {
	select {
	case tb := <-download.DQueue.Queue:
		torrent := tb.Torrent
		bangumi := tb.Bangumi
		download.DQueue.InQueue.Delete(torrent.Link)

		slog.Debug("[download task] 开始下载种子", "Name", torrent.Name)
		guids, err := download.Client.Add(ctx, torrent.Link, genSavePath(bangumi))
		if err != nil {
			slog.Error("[download task] 下载种子失败", "Name", torrent.Name, "error", err)
			// 重新加入队列，稍后重试
			download.DQueue.Add(ctx, torrent, bangumi)
			return nil
		}
		slog.Debug("[download task] 下载种子成功", "Name", torrent.Name, "GUIDs", guids)

		// 创建任务提交到 taskrunner
		task := model.NewTask(torrent, bangumi)
		task.Guids = guids
		t.runner.Submit(task)
	default:
		// 队列为空，跳过
	}
	return nil
}

func genSavePath(bangumi *model.Bangumi) string {
	folder := bangumi.OfficialTitle
	if bangumi.Year != "" {
		folder += " (" + bangumi.Year + ")"
	}
	season := "Season " + strconv.Itoa(bangumi.Season)
	fp := filepath.Join(folder, season)
	return fp
}
