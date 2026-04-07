package rename

import (
	"context"
	"log/slog"
	"path/filepath"
	"strconv"

	"goto-bangumi/internal/download"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/notification"
	"goto-bangumi/internal/parser"
)

// Rename renames (moves) oldpath to newpath.
// 主要的流程为:
// 1. 拿到 torrent 和 bangumi 信息
// 2. 通过 torrent 的guid 信息去拿到种子里面有什么文件
// 3. 遍历文件,拿到要重命名的文件路径
// 4. 生成新的文件路径
// 调用 downloadClient 实现重命名
// 如果重命名失败,则返回错误
// 如果成功, 则把把数据库内 torrent 的 状态更新为已重命名

var renameConfig = &model.BangumiRenameConfig{}

func Init(cfg *model.BangumiRenameConfig) {
	renameConfig = cfg
}

func GetBangumi(ctx context.Context, torrent *model.Torrent) (*model.Bangumi, error) {
	return getBangumi(ctx, torrent)
}

func Rename(ctx context.Context, torrent *model.Torrent, bangumi *model.Bangumi) {
	// 如果 bangumi 为空, 则从 torrent 中获取 bangumi 信息
	if bangumi == nil {
		var err error
		bangumi, err = getBangumi(ctx, torrent)
		if err != nil {
			return
		}
	}
	fileList, err := download.Client.GetTorrentFiles(ctx, torrent.DownloadUID)
	if err != nil {
		return
	}

	for _, filePath := range fileList {
		// 从 file_path 中提取出文件名, 通过 filepath
		torrentName := filepath.Base(filePath)
		// 跳过 0.5 集的文件
		if parser.IsPoint5(torrentName) {
			slog.Debug("[rename] Skip renaming for 0.5 episode file", "file", torrentName)
			continue
		}
		metaInfo, newPath := GenPath(torrentName, bangumi)
		if newPath == filePath {
			slog.Debug("[rename] File path is the same, no need to rename", "path", filePath)
			continue
		}

		// 也不用想着要加速什么的, 慢慢来就好了, 主要的还是 api 调用的时间
		// err := rename(ctx, torrent.DownloadUID, filePath, newPath)
		if err := download.Client.Rename(ctx, torrent.DownloadUID, filePath, newPath); err != nil {
			slog.Error("[rename] Failed to rename file", "oldpath", filePath, "newpath", newPath, "error", err)
			return
		}

		// 发送改名成功通知
		Nclient := notification.NotificationClient
		msg := &model.Message{
			Title:      bangumi.OfficialTitle,
			Season:     strconv.Itoa(bangumi.Season),
			Episode:    strconv.Itoa(metaInfo.Episode),
			PosterLink: bangumi.PosterLink,
		}
		Nclient.Send(ctx, msg)
	}
}

