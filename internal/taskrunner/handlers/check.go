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

// NewCheckHandler 创建检查处理器，验证下载是否成功添加到下载器
func NewCheckHandler(db *database.DB) taskrunner.PhaseFunc {
	return func(ctx context.Context, task *model.Task) taskrunner.PhaseResult {
		for _, guid := range task.Guids {
			trueID, err := download.Client.Check(ctx, guid)

			// GUID 没找到，试下一个
			if apperrors.IsKeyError(err) {
				continue
			}

			if err != nil {
				slog.Error("[check handler] 检查下载失败", "error", err)
				return taskrunner.PhaseResult{Err: err}
			}

			// 找到了真实 ID
			if trueID != "" {
				task.Torrent.DownloadUID = trueID

				if err := db.AddTorrentDUID(ctx, task.Torrent.Link, trueID); err != nil {
					slog.Error("[check handler] 更新 Torrent DUID 失败", "error", err)
					return taskrunner.PhaseResult{Err: err}
				}

				slog.Debug("[check handler] 获取到真实 DUID",
					"torrent", task.Torrent.Name, "duid", trueID)

				return taskrunner.PhaseResult{} // 成功
			}
		}

		return taskrunner.PhaseResult{Err: errors.New("no valid hash found")}
	}
}
