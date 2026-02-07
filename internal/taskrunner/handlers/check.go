package handlers

import (
	"context"
	"errors"
	"log/slog"

	"goto-bangumi/internal/apperrors"
	"goto-bangumi/internal/database"
	"goto-bangumi/internal/download"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/taskrunner"
)

// CheckHandler 检查下载是否成功添加到下载器
type CheckHandler struct{}

func (h *CheckHandler) Handle(ctx context.Context, task *model.Task) taskrunner.HandlerResult {
	for _, guid := range task.Guids {
		trueID, err := download.Client.Check(ctx, guid)

		// 如果表明没有找到，尝试下一个 guid
		if apperrors.IsKeyError(err) {
			continue
		}

		// 如果是登录错误，直接失败不重试
		if apperrors.IsDownloadLoginError(err) {
			slog.Error("[check handler] 检查下载失败，登录错误", "error", err)
			return taskrunner.HandlerResult{
				Error:       err,
				ShouldRetry: false,
			}
		}

		// 其他错误，可重试
		if err != nil {
			slog.Error("[check handler] 检查下载失败", "error", err)
			return taskrunner.HandlerResult{
				Error:       err,
				ShouldRetry: true,
			}
		}

		// 找到了真实 ID
		if trueID != "" {
			task.Torrent.DownloadUID = trueID

			// 存入数据库
			if err := database.GetDB().AddTorrentDUID(ctx,task.Torrent.Link, trueID); err != nil {
				slog.Error("[check handler] 更新 Torrent DUID 失败", "error", err)
				return taskrunner.HandlerResult{
					Error:       err,
					ShouldRetry: true,
				}
			}

			slog.Debug("[check handler] 获取到真实 DUID", "torrent", task.Torrent.Name, "duid", trueID)

			return taskrunner.HandlerResult{
				NextPhase: model.PhaseDownloading,
			}
		}
	}

	// 所有 guid 都没找到，重试
	return taskrunner.HandlerResult{
		Error:       errors.New("no valid hash found"),
		ShouldRetry: true,
	}
}
