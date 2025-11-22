package download

// 实现Add之前的处理, 让加入和Download 解耦
// 一个的下载队列, 负责管理下载任务的加入和调度
// 另一个就是下载客户端, 负责实际的下载操作

import (
	"context"
	"log/slog"
	"sync"

	"goto-bangumi/internal/model"
)

var dqueue chan *model.TorrentBangumi = make(chan *model.TorrentBangumi, 100)

type DownloadQueue struct {
	Queue chan *model.TorrentBangumi
	// 使用 sync.Map 跟踪队列中的 torrent URL，防止重复添加
	InQueue sync.Map // key: torrent.URL (string), value: bool
}

var DQueue *DownloadQueue = &DownloadQueue{
	Queue:   dqueue,
	InQueue: sync.Map{},
}

func (dq *DownloadQueue) Add(ctx context.Context, torrent *model.Torrent, bangumi *model.Bangumi) {
	select {
	case <-Client.LoginError:
		slog.Warn("下载客户端登陆任务退出，无法添加种子", "Name", torrent.Name)
		return
	default:
	}
	// 看看是不是在登陆, 登陆中的话就等登陆完成
	err := Client.EnsureLogin(ctx)
	if err != nil {
		slog.Warn("[DownloadQueue] 下载客户端登陆失败，无法添加种子", "Name", torrent.Name, "error", err)
		return
	}
	// 检查该 torrent URL 是否已在队列中排队
	if _, exists := dq.InQueue.Load(torrent.URL); exists {
		slog.Debug("种子已在下载队列中，跳过添加", "Name", torrent.Name, "URL", torrent.URL)
		return
	}

	tb := &model.TorrentBangumi{
		Bangumi: bangumi,
		Torrent: torrent,
	}
	select {
	case <-ctx.Done():
		slog.Warn("下载队列已关闭，无法添加种子", "Name", torrent.Name)
	case dq.Queue <- tb:
		// 成功加入队列后，标记该 URL 正在队列中
		dq.InQueue.Store(torrent.URL, true)
		slog.Debug("种子已加入下载队列", "Name", torrent.Name, "URL", torrent.URL)
	}
}

func (dq *DownloadQueue) Clear() {
	// 清理队列和标记
	for {
		select {
		case tb := <-dq.Queue:
			// 清理队列时也要移除对应的标记
			dq.InQueue.Delete(tb.Torrent.URL)
		default:
			return
		}
	}
}
