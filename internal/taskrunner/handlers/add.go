package handlers

import (
	"context"
	"log/slog"
	"path/filepath"
	"strconv"
	"time"

	"goto-bangumi/internal/download"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/taskrunner"
)

// NewAddHandler 创建添加下载处理器，将种子添加到下载器
func NewAddHandler() taskrunner.PhaseFunc {
	return func(ctx context.Context, task *model.Task) taskrunner.PhaseResult {
		savePath := genSavePath(task.Bangumi)
		guids, err := download.Client.Add(ctx, task.Torrent.Link, savePath)
		if err != nil {
			slog.Warn("[add handler] 添加下载失败，稍后重试",
				"torrent", task.Torrent.Name, "error", err)
			return taskrunner.PhaseResult{PollAfter: 5 * time.Second}
		}

		task.Guids = guids
		slog.Debug("[add handler] 添加下载成功",
			"torrent", task.Torrent.Name, "guids", guids)
		return taskrunner.PhaseResult{}
	}
}

// genSavePath 根据番剧信息生成保存路径
func genSavePath(bangumi *model.Bangumi) string {
	folder := bangumi.OfficialTitle
	if bangumi.Year != "" {
		folder += " (" + bangumi.Year + ")"
	}
	season := "Season " + strconv.Itoa(bangumi.Season)
	return filepath.Join(folder, season)
}
