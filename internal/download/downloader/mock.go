package downloader

import (
	"context"

	"goto-bangumi/internal/model"
)

// MockDownloader 模拟下载器实现，用于测试
type MockDownloader struct {
	config      *model.DownloaderConfig
	APIInterval int
}

// NewMockDownloader 创建新的模拟下载器
func NewMockDownloader() *MockDownloader {
	return &MockDownloader{
		APIInterval: 100,
	}
}

// Init 初始化下载器
func (d *MockDownloader) Init(config *model.DownloaderConfig) error {
	d.config = config
	return nil
}

// Auth 认证登录
func (d *MockDownloader) Auth(ctx context.Context) (bool, error) {
	return true, nil
}

// Logout 登出
func (d *MockDownloader) Logout(ctx context.Context) (bool, error) {
	return true, nil
}

// GetTorrentFiles 获取种子文件列表
func (d *MockDownloader) GetTorrentFiles(ctx context.Context, hash string) ([]string, error) {
	if files, ok := MockFiles[hash]; ok {
		return files, nil
	}
	return []string{}, nil
}

// GetTorrentInfo 获取单个种子详细信息
func (d *MockDownloader) GetTorrentInfo(ctx context.Context, hash string) (*model.TorrentDownloadInfo, error) {
	if info, ok := MockTorrentInfos[hash]; ok {
		info.SavePath = d.config.SavePath + "/" + info.SavePath
		return info, nil
	}
	return nil, nil
}

// TorrentsInfo 获取种子信息列表
func (d *MockDownloader) TorrentsInfo(ctx context.Context, statusFilter, category string, tag *string, limit int) ([]map[string]any, error) {
	torrents := []map[string]any{
		{
			"hash":          "abc123def456",
			"name":          "[SubGroup] Anime Title - 01 [1080p]",
			"state":         "completed",
			"progress":      1.0,
			"size":          1073741824,
			"completion_on": 1704067200,
			"save_path":     "/downloads/anime",
			"category":      "GotoBangumi",
		},
		{
			"hash":          "xyz789ghi012",
			"name":          "[SubGroup] Anime Title - 02 [1080p]",
			"state":         "downloading",
			"progress":      0.5,
			"size":          1073741824,
			"completion_on": -1,
			"save_path":     "/downloads/anime",
			"category":      "GotoBangumi",
		},
	}

	if limit > 0 && limit < len(torrents) {
		return torrents[:limit], nil
	}
	return torrents, nil
}

// CheckHash 检查种子是否存在
func (d *MockDownloader) CheckHash(ctx context.Context, hash string) (string, error) {
	return hash, nil
}

// Add 添加种子
func (d *MockDownloader) Add(ctx context.Context, torrentInfo *model.TorrentInfo, savePath string) ([]string, error) {
	hashes := make([]string, 0, 2)
	if torrentInfo.InfoHashV1 != "" {
		hashes = append(hashes, torrentInfo.InfoHashV1)
	}
	if torrentInfo.InfoHashV2 != "" {
		v2Hash := torrentInfo.InfoHashV2
		if len(v2Hash) > 40 {
			v2Hash = v2Hash[:40]
		}
		hashes = append(hashes, v2Hash)
	}
	return hashes, nil
}

// Delete 删除种子
func (d *MockDownloader) Delete(ctx context.Context, hashes []string) (bool, error) {
	return true, nil
}

// Rename 重命名种子文件
func (d *MockDownloader) Rename(ctx context.Context, torrentHash, oldPath, newPath string) (bool, error) {
	return true, nil
}

// Move 移动种子到新位置
func (d *MockDownloader) Move(ctx context.Context, hashes []string, newLocation string) (bool, error) {
	return true, nil
}

// GetInterval 获取 API 调用间隔
func (d *MockDownloader) GetInterval() int {
	return d.APIInterval
}
