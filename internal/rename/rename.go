package rename

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/download"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/network"
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

// Renamer 封装重命名相关操作
type Renamer struct {
	db         *database.DB
	downloader *download.DownloadClient
}

// New 创建 Renamer 实例
func New(db *database.DB, dl *download.DownloadClient) *Renamer {
	return &Renamer{db: db, downloader: dl}
}

func (r *Renamer) GetBangumi(ctx context.Context, torrent *model.Torrent) (*model.Bangumi, error) {
	return r.getBangumi(ctx, torrent)
}

func (r *Renamer) Rename(ctx context.Context, torrent *model.Torrent, bangumi *model.Bangumi) {
	// 如果 bangumi 为空, 则从 torrent 中获取 bangumi 信息
	if bangumi == nil {
		var err error
		bangumi, err = r.getBangumi(ctx, torrent)
		if err != nil {
			return
		}
	}
	fileList, err := r.downloader.GetTorrentFiles(ctx, torrent.DownloadUID)
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
		if err := r.downloader.Rename(ctx, torrent.DownloadUID, filePath, newPath); err != nil {
			slog.Error("[rename] Failed to rename file", "oldpath", filePath, "newpath", newPath, "error", err)
			return
		}

		// 发送改名成功通知
		text := fmt.Sprintf("番剧名称：%s\n季度：第%d季\n更新集数：第%d集",
			bangumi.OfficialTitle, bangumi.Season, metaInfo.Episode)

		var image []byte
		if bangumi.PosterLink != "" {
			image, err = network.LoadImage(ctx, bangumi.PosterLink)
			if err != nil {
				slog.Error("[rename] Failed to download poster", "error", err)
			}
		}

		notification.NotificationClient.Send(ctx, &notification.Message{
			Text:  text,
			Image: image,
		})
	}
}

