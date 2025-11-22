package subscribe

import (
	"context"
	"log/slog"
	"time"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/download"
	"goto-bangumi/internal/eventbus"
	"goto-bangumi/internal/model"
)

// 用来检查下载是否完成了, 如果完成了就发出下载完成事件

type checkDownloadingService struct {
	bus eventbus.EventBus
}

// calculateEta 根据当前下载进度计算预计剩余时间（秒）
func calculateEta(eta int64) int {
	// 已完成或无效
	if eta <= 0 {
		return 10
	}
	// 小于1分钟，每10秒检查
	if eta < 60 {
		return 10
	}
	// 小于5分钟，每30秒检查
	if eta < 300 {
		return 30
	}
	// 小于30分钟，每2分钟检查
	if eta < 1800 {
		return 120
	}
	return 300
}

// handleDownloadingCheck 处理单个下载检查事件
func (cds *checkDownloadingService) handleDownloadingCheck(ctx context.Context, data model.DownloadingCheckEvent) {
	// 如果是第一次检查（StartTime 为零值），设置开始时间
	if data.StartTime.IsZero() {
		data.StartTime = time.Now()
	}

	// 计算已经检查了多久
	elapsed := time.Since(data.StartTime)
	// 如果已经检查超过 4 小时，标记为异常
	if elapsed > 4*time.Hour {
		slog.Warn("[check downloading service] 检查下载超过4小时，标记为不可下载",
			"hash", data.Torrent.DownloadUID, "elapsed", elapsed)
		// 更新状态为 4（异常/手动停止下载）
		data.Torrent.Downloaded = 4
		db := database.GetDB()
		if err := db.UpdateTorrent(data.Torrent); err != nil {
			slog.Error("[check downloading service] 更新种子状态失败", "error", err)
		}
		return
	}

	// 获取种子信息
	info, err := download.Client.GetTorrentInfo(ctx, data.Torrent.DownloadUID)
	if err != nil {
		slog.Error("[check downloading service] 获取种子信息失败", "error", err, "hash", data.Torrent.DownloadUID)
		return
	}
	if info == nil {
		slog.Warn("[check downloading service] 种子不存在", "hash", data.Torrent.DownloadUID)
		return
	}

	eta := int64(info.ETA)

	// 使用 calculate_eta 计算等待时间并睡眠
	waitTime := calculateEta(eta)
	time.Sleep(time.Duration(waitTime) * time.Second)

	// 睡眠后再次检查下载状态
	info2, err := download.Client.GetTorrentInfo(ctx, data.Torrent.DownloadUID)
	if err != nil {
		slog.Error("[check downloading service] 检查下载状态失败", "error", err, "hash", data.Torrent.DownloadUID)
		return
	}
	if info2 == nil {
		slog.Warn("[check downloading service] 种子不存在", "hash", data.Torrent.DownloadUID)
		return
	}

	// 检查是否下载完成（Completed > 0 表示已完成，为 Unix 时间戳）
	if info2.Completed > 0 {
		// 下载完成，更新状态为 2
		data.Torrent.Downloaded = 2
		db := database.GetDB()
		if err := db.UpdateTorrent(data.Torrent); err != nil {
			slog.Error("[check downloading service] 更新种子状态失败", "error", err)
			return
		}

		// 发布重命名事件
		cds.bus.Publish(ctx, model.RenameEvent{
			Torrent: data.Torrent,
			Bangumi: data.Bangumi,
		})
	} else {
		// 未完成，重新发布检查事件继续循环（携带 StartTime）
		cds.bus.Publish(ctx, model.DownloadingCheckEvent{
			Torrent:   data.Torrent,
			Bangumi:   data.Bangumi,
			StartTime: data.StartTime, // 保持原始开始时间
		})
	}
}

func (cds *checkDownloadingService) Start(ctx context.Context) {
	ch, unsubscribe := eventbus.Subscribe[model.DownloadingCheckEvent](cds.bus, ctx, 100)
	defer unsubscribe()

	for event := range ch {
		go cds.handleDownloadingCheck(ctx, event)
	}
}
