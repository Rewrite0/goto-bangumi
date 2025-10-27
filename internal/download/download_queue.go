package download

// 实现Add之前的处理, 让加入和Download 解耦
// 一个的下载队列, 负责管理下载任务的加入和调度
// 另一个就是下载客户端, 负责实际的下载操作

import (
	"context"
	"log/slog"
	"time"

	"goto-bangumi/internal/model"
)

var dqueue chan *model.Torrent = make(chan *model.Torrent, 100)

type DownloadQueue struct {
	Queue chan *model.Torrent
}

var DQueue *DownloadQueue = &DownloadQueue{
	Queue: dqueue,
}

func (dq *DownloadQueue) Add(ctx context.Context, torrent *model.Torrent) {
	select {
	case <-ctx.Done():
		slog.Warn("下载队列已关闭，无法添加种子", "Name", torrent.Name)
	case dq.Queue <- torrent:
	}
}

func (dq *DownloadQueue) Clear() {
	// 清理队列
	for {
		select {
		case <-dq.Queue:
		default:
			return
		}
	}
}

func DownloadLoop(ctx context.Context, client *DownloadClient) {
	for {
		select {
		case torrent := <-DQueue.Queue:
			slog.Info("开始下载种子", "Name", torrent.Name)
			//TODO: 生成保存路径
			guid, err := client.Add(ctx, torrent.URL, "")
			if err != nil {
				slog.Error("下载种子失败", "Name", torrent.Name, "error", err)
				continue
			}
			slog.Info("下载种子成功", "Name", torrent.Name, "GUID", guid)
			time.Sleep(5 * time.Second) // 避免过快添加任务
		case <-ctx.Done():
			slog.Info("下载队列退出")
			// 清理队列
			dqueue = make(chan *model.Torrent, 100)
			return
		}
	}
}
