package handlers

import (
	"context"
	"log/slog"
	"path/filepath"
	"strconv"
	"time"

	"goto-bangumi/internal/apperrors"
	"goto-bangumi/internal/download"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/taskrunner"
)

// NewAddHandler 创建添加下载处理器，将种子添加到下载器
func NewAddHandler(dl *download.DownloadClient) taskrunner.PhaseFunc {
	return func(ctx context.Context, task *model.Task) taskrunner.PhaseResult {
		savePath := genSavePath(task.Bangumi)
		guids, err := dl.Add(ctx, task.Torrent.Link, savePath)
		if err != nil {
			slog.Warn("[add handler] 添加下载失败，稍后重试",
				"torrent", task.Torrent.Name, "error", err)
			// TODO: 不应一直重试, 一是要有次数的限制, 二是要看是什么错误
			if apperrors.IsNetworkError(err) {
				return taskrunner.PhaseResult{PollAfter: 5 * time.Second}
			}
			return taskrunner.PhaseResult{Err: err}
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
