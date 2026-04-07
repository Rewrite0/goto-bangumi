package handlers

import (
	"context"
	"log/slog"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/rename"
	"goto-bangumi/internal/taskrunner"
)

// NewRenameHandler 创建重命名处理器
func NewRenameHandler() taskrunner.PhaseFunc {
	return func(ctx context.Context, task *model.Task) taskrunner.PhaseResult {
		slog.Info("[rename handler] 开始重命名",
			"torrent", task.Torrent.Name,
			"bangumi", task.Bangumi.OfficialTitle)

		rename.Rename(ctx, task.Torrent, task.Bangumi)

		if err := database.GetDB().TorrentRenamed(ctx, task.Torrent.Link); err != nil {
			slog.Error("[rename handler] 更新种子重命名状态失败",
				"error", err, "link", task.Torrent.Link)
			return taskrunner.PhaseResult{Err: err}
		}

		slog.Info("[rename handler] 重命名完成", "torrent", task.Torrent.Name)
		return taskrunner.PhaseResult{} // 成功
	}
}
