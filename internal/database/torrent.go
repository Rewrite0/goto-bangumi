package database

import (
	"context"
	"log/slog"

	"goto-bangumi/internal/model"
)

// ============ Torrent 相关方法 ============

// CreateTorrent 创建种子
func (db *DB) CreateTorrent(ctx context.Context, torrent *model.Torrent) error {
	// 先检查是否存在相同 link 的种子
	var existing model.Torrent
	err := db.WithContext(ctx).Where("link = ?", torrent.Link).First(&existing).Error
	if err == nil {
		slog.Info("[database] 种子已存在，跳过创建", "link", torrent.Link)
		return nil
	}
	// 创建新种子
	err = db.WithContext(ctx).Create(torrent).Error
	return err
}

// AddTorrentDownload 种子标记为已下载
func (db *DB) AddTorrentDownload(ctx context.Context, link string) error {
	t := model.Torrent{}
	err := db.WithContext(ctx).Where("link = ?", link).First(&t).Error
	if err != nil {
		slog.Error("[database] 标记种子已下载失败，未找到种子", "link", link, "error", err)
		return err
	}
	t.Downloaded = model.DownloadDone
	err = db.WithContext(ctx).Save(&t).Error
	return err
}

func (db *DB) AddTorrentError(ctx context.Context, link string) error {
	t := model.Torrent{}
	err := db.WithContext(ctx).Where("link = ?", link).First(&t).Error
	if err != nil {
		slog.Error("[database] 标记种子下载出错失败，未找到种子", "link", link, "error", err)
		return err
	}
	t.Downloaded = model.DownloadError
	err = db.WithContext(ctx).Save(&t).Error
	return err
}

func (db *DB) TorrentRenamed(ctx context.Context, link string) error {
	t := model.Torrent{}
	err := db.WithContext(ctx).Where("link = ?", link).First(&t).Error
	if err != nil {
		slog.Error("[database] 标记种子已重命名失败，未找到种子", "link", link, "error", err)
		return err
	}
	t.Renamed = true
	err = db.WithContext(ctx).Save(&t).Error
	return err
}

// DeleteTorrent 删除种子
func (db *DB) DeleteTorrent(ctx context.Context, link string) error {
	return db.WithContext(ctx).Where("link = ?", link).Delete(&model.Torrent{}).Error
}

// AddTorrentDUID 为种子添加下载 UID
func (db *DB) AddTorrentDUID(ctx context.Context, link string, guid string) error {
	t := model.Torrent{}
	err := db.WithContext(ctx).Where("link = ?", link).First(&t).Error
	if err != nil {
		slog.Error("[database] 添加种子 UID 失败，未找到种子", "link", link, "error", err)
		return err
	}
	t.DownloadUID = guid
	// 标记为已发送到下载器
	t.Downloaded = model.DownloadSending
	err = db.WithContext(ctx).Save(&t).Error
	return err
}
