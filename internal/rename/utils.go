package rename

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"goto-bangumi/internal/download"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/parser"
)

func (r *Renamer) getBangumi(ctx context.Context, torrent *model.Torrent) (*model.Bangumi, error) {
	// 从 download 中拿到下载文件的目录信息
	downloadInfo, err := download.Client.GetTorrentInfo(ctx, torrent.DownloadUID)
	if err != nil {
		slog.Error("[rename] Failed to get torrent download info", "name", torrent.Name, "error", err)
		return nil, err
	}
	savePath := downloadInfo.SavePath
	// 从 savePath 提取出 bangumi 的名字和季度 以及 可能存在的年份 组成为 savePath/BangumiName (Year)/Season \d
	// 首先提取一个相对路径, 拿到最后的 BangumiName (Year)/Season \d, 以 download.Client.SavePath 为基准
	relativePath, err := filepath.Rel(download.Client.SavePath, savePath)
	if err != nil {
		slog.Error("[rename] Failed to get relative path", "name", torrent.Name, "path", savePath, "error", err)
		return nil, err
	}
	pathInfo := parser.ParsePath(relativePath)
	// 不解析非标准的路径
	if pathInfo == nil {
		slog.Error("[rename] Failed to parse path info", "name", torrent.Name, "relativePath", relativePath)
		return nil, fmt.Errorf("failed to parse path info")
	}
	// 去数据库中查找 bangumi 信息
	bangumi, err := r.db.GetBangumiByOfficialTitle(pathInfo.BangumiName)
	if err != nil {
		slog.Debug("[rename] Failed to get bangumi from database", "name", torrent.Name, "bangumiName", pathInfo.BangumiName, "error", err)
		// 如果没有找到的话,就新建一个 bangumi
		bangumi = &model.Bangumi{
			OfficialTitle: pathInfo.BangumiName,
			Season:        pathInfo.SeasonNumber,
			Year:          pathInfo.Year,
		}
	}
	return bangumi, nil
}

// GenPath 生成新的文件路径,形如 败犬女主太多了 (2024) S01E02 - Ani.mp4
func GenPath(torrentName string, bangumi *model.Bangumi) (*model.EpisodeMetadata, string) {
	metaInfo := parser.NewTitleMetaParse().ParseEpisode(torrentName)
	episode := metaInfo.Episode
	if episode == -1 {
		slog.Error("[rename] Failed to parse episode from torrent name", "torrentName", torrentName)
		return nil, ""
	}

	// offset, 默认是0
	episode += bangumi.Offset

	// 获取文件扩展名
	ext := filepath.Ext(torrentName)

	// 构建基本路径: OfficialTitle
	newPath := bangumi.OfficialTitle

	// 添加年份 (如果配置启用且存在)
	if renameConfig.Year && bangumi.Year != "" {
		newPath += fmt.Sprintf(" (%s)", bangumi.Year)
	}

	// 添加季度和集数: S01E02
	newPath += fmt.Sprintf(" S%02dE%02d", bangumi.Season, episode)

	// 添加字幕组信息 (如果配置启用且存在)
	if renameConfig.Group && metaInfo.Group != "" {
		newPath += fmt.Sprintf(" - %s", metaInfo.Group)
	}

	// 添加文件扩展名
	newPath += ext
	// TODO: 字幕文件还要加 chs, cht 等标识
	return metaInfo, newPath
}
