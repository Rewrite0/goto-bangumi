package handlers

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/download"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/taskrunner"
)

// ProgressHandler 检查下载进度
type ProgressHandler struct{}

func (h *ProgressHandler) Handle(ctx context.Context, task *model.Task) taskrunner.HandlerResult {
	// 检查是否超时（4小时）
	if time.Since(task.StartTime) > 4*time.Hour {
		slog.Warn("[progress handler] 下载超过4小时，标记为异常",
			"torrent", task.Torrent.Name,
			"duid", task.DownloadUID,
			"elapsed", time.Since(task.StartTime))

		// 更新数据库状态为异常
		task.Torrent.Downloaded = 4
		if err := database.GetDB().AddTorrentError(task.Torrent.Link); err != nil {
			slog.Error("[progress handler] 更新种子状态失败", "error", err)
		}

		return taskrunner.HandlerResult{
			Error:       errors.New("download timeout after 4 hours"),
			ShouldRetry: false,
		}
	}

	// 获取种子信息
	info, err := download.Client.GetTorrentInfo(ctx, task.DownloadUID)
	if err != nil {
		slog.Error("[progress handler] 获取种子信息失败", "error", err, "duid", task.DownloadUID)
		return taskrunner.HandlerResult{
			Error:       err,
			ShouldRetry: true,
		}
	}

	if info == nil {
		slog.Warn("[progress handler] 种子不存在", "duid", task.DownloadUID)
		return taskrunner.HandlerResult{
			Error:       errors.New("torrent not found"),
			ShouldRetry: false,
		}
	}

	// 检查是否下载完成（Completed > 0 表示已完成，为 Unix 时间戳）
	if info.Completed > 0 {
		task.Torrent.Downloaded = 2
		if err := database.GetDB().AddTorrentDownload(task.Torrent.Link); err != nil {
			slog.Error("[progress handler] 更新种子状态失败", "error", err)
			return taskrunner.HandlerResult{
				Error:       err,
				ShouldRetry: true,
			}
		}

		slog.Info("[progress handler] 下载完成", "torrent", task.Torrent.Name)

		return taskrunner.HandlerResult{
			NextPhase: model.PhaseRenaming,
		}
	}

	// 更新 ETA
	task.LastETA = info.ETA

	// 未完成，保持当前阶段，等待下次检查
	// ScheduleAfter 由 ETAHandler 设置
	return taskrunner.HandlerResult{
		NextPhase: model.PhaseDownloading,
	}
}

// ETAHandler 根据 ETA 计算下次检查间隔
type ETAHandler struct{}

func (h *ETAHandler) Handle(ctx context.Context, task *model.Task) taskrunner.HandlerResult {
	// 如果任务已进入下一阶段（Renaming），不需要设置延迟
	if task.Phase != model.PhaseDownloading {
		return taskrunner.HandlerResult{}
	}

	interval := calculateEta(int64(task.LastETA))

	slog.Debug("[eta handler] 设置检查间隔",
		"torrent", task.Torrent.Name,
		"eta", task.LastETA,
		"interval", interval)

	return taskrunner.HandlerResult{
		NextPhase:     model.PhaseDownloading,
		ScheduleAfter: time.Duration(interval) * time.Second,
	}
}

// calculateEta 根据 ETA 计算检查间隔（秒）
func calculateEta(eta int64) int {
	if eta <= 0 {
		return 10
	}
	if eta < 60 {
		return 10
	}
	if eta < 300 {
		return 30
	}
	if eta < 1800 {
		return 120
	}
	return 300
}
