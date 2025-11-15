package download

// 实现Add之前的处理, 让加入和Download 解耦
// 一个的下载队列, 负责管理下载任务的加入和调度
// 另一个就是下载客户端, 负责实际的下载操作

import (
	"context"
	"log/slog"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"goto-bangumi/internal/model"
)

var dqueue chan *model.TorrentBangumi = make(chan *model.TorrentBangumi, 100)

type DownloadQueue struct {
	Queue chan *model.TorrentBangumi
	// 使用 sync.Map 跟踪队列中的 torrent URL，防止重复添加
	inQueue sync.Map // key: torrent.URL (string), value: bool
}

var DQueue *DownloadQueue = &DownloadQueue{
	Queue:   dqueue,
	inQueue: sync.Map{},
}

func (dq *DownloadQueue) Add(ctx context.Context, torrent *model.Torrent, bangumi *model.Bangumi) {
	select {
	case <-Client.loginError:
		slog.Warn("下载客户端登陆任务退出，无法添加种子", "Name", torrent.Name)
		return
	default:
	}
	// 检查该 torrent URL 是否已在队列中排队
	if _, exists := dq.inQueue.Load(torrent.URL); exists {
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
		dq.inQueue.Store(torrent.URL, true)
		slog.Debug("种子已加入下载队列", "Name", torrent.Name, "URL", torrent.URL)
	}
}

func (dq *DownloadQueue) Clear() {
	// 清理队列和标记
	for {
		select {
		case tb := <-dq.Queue:
			// 清理队列时也要移除对应的标记
			dq.inQueue.Delete(tb.Torrent.URL)
		default:
			return
		}
	}
}

func DownloadLoop(ctx context.Context, client *DownloadClient) {
	for {
		// TODO: 没有登陆的时候,不要取出队列
		select {
		case <-Client.loginError:
			slog.Warn("下载队列退出，下载客户端已停止工作")
			DQueue.Clear()
			return
		case tb := <-DQueue.Queue:
			torrent := tb.Torrent
			bangumi := tb.Bangumi
			// 从队列取出后，立即移除标记，表示该 URL 不再在队列中
			DQueue.inQueue.Delete(torrent.URL)

			slog.Info("开始下载种子", "Name", torrent.Name)
			guid, err := client.Add(ctx, torrent.URL, genSavePath(bangumi))
			if err != nil {
				slog.Error("下载种子失败", "Name", torrent.Name, "error", err)
				continue
			}
			slog.Info("下载种子成功", "Name", torrent.Name, "GUID", guid)
			time.Sleep(5 * time.Second) // 避免过快添加任务
		case <-ctx.Done():
			slog.Info("下载队列退出")
			// 清理队列
			DQueue.Clear()
			return
		}
	}
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
