package task

import (
	"context"
	"log/slog"
	"path/filepath"
	"strconv"
	"time"

	"goto-bangumi/internal/download"
	"goto-bangumi/internal/model"
)

// DownloadTask 下载任务
type DownloadTask struct {
	interval time.Duration
	enabled  bool
}

// NewDownloadTask 创建下载任务
func NewDownloadTask() *DownloadTask {
	return &DownloadTask{
		interval: 5 * time.Second,
		enabled:  true,
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
		download.DQueue.InQueue.Delete(torrent.URL)

		slog.Debug("[download task] 开始下载种子", "Name", torrent.Name)
		guid, err := download.Client.Add(ctx, torrent.URL, genSavePath(bangumi))
		if err != nil {
			slog.Error("[download task] 下载种子失败", "Name", torrent.Name, "error", err)
			// 重新加入队列，稍后重试
			download.DQueue.Add(ctx, torrent, bangumi)
		}
		slog.Debug("下载种子成功", "Name", torrent.Name, "GUID", guid)
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
