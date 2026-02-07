package handlers

import (
	"context"
	"log/slog"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/rename"
	"goto-bangumi/internal/taskrunner"
)

// RenameHandler 重命名处理器
type RenameHandler struct{}

func (h *RenameHandler) Handle(ctx context.Context, task *model.Task) taskrunner.HandlerResult {
	slog.Info("[rename handler] 开始重命名",
		"torrent", task.Torrent.Name,
		"bangumi", task.Bangumi.OfficialTitle)

	// 调用 rename 模块进行重命名
	rename.Rename(ctx, task.Torrent, task.Bangumi)

	// 更新数据库状态为已重命名
	if err := database.GetDB().TorrentRenamed(task.Torrent.Link); err != nil {
		slog.Error("[rename handler] 更新种子重命名状态失败", "error", err, "link", task.Torrent.Link)
		return taskrunner.HandlerResult{
			Error:       err,
			ShouldRetry: true,
		}
	}

	slog.Info("[rename handler] 重命名完成", "torrent", task.Torrent.Name)

	return taskrunner.HandlerResult{
		NextPhase: model.PhaseCompleted,
	}
}
