package downloader

import (
	"fmt"
	"strings"

	"goto-bangumi/internal/model"
)

// BaseDownloader 定义下载器的基础接口
// 主要实现以下功能:
// 1. renamer,用以实现重命名功能 hash, old name, new name
// 2. move, 用以实现改数据后的转移
// 3. add, 用以加种子
// 4. get_info, 获取当前所有的有关种子
// 5. auth, 用以登陆
// 6. check_host, 用以检查连通性
// 7. logout,用以登出
type BaseDownloader interface {
	// Init 初始化下载器
	Init(config *model.DownloaderConfig) error

	// Auth 用户认证
	Auth() (bool, error)

	// CheckHost 检查主机连通性
	// CheckHost() (bool, error)

	// Logout 登出
	Logout() (bool, error)

	// GetTorrentFiles 获取种子的所有文件列表
	GetTorrentFiles(hash string) ([]string, error)

	// GetTorrentInfo 获取单个种子的信息
	GetTorrentInfo(hash string)(*model.TorrentDownloadInfo,error)

	// TorrentsInfo 获取种子信息列表
	TorrentsInfo(statusFilter, category string, tag *string, limit int) ([]map[string]interface{}, error)

	// Rename 重命名种子文件
	Rename(torrentHash, oldPath, newPath string) (bool, error)

	// Move 移动种子到新位置
	Move(hashes []string, newLocation string) (bool, error)

	// Add 添加种子
	Add(torrentInfo *model.TorrentInfo, savePath string) (string, error)

	// CheckHash 检查种子是否存在，返回真实的hash
	CheckHash(hash string) (string, error)

	// Delete 删除种子
	Delete(hashes []string) (bool, error)

	// GetInterval 获取下载器轮询间隔时间，单位ms
	GetInterval() int
}

// NewDownloader 创建下载器实例
// 根据 downloaderType 动态选择具体的下载器实现
// 支持的类型: "qbittorrent", "qb"
func NewDownloader(downloaderType string) (BaseDownloader, error) {
	if downloaderType == "" {
		return nil, fmt.Errorf("下载器类型不能为空")
	}

	var d BaseDownloader

	// 根据类型选择具体的下载器实现
	switch strings.ToLower(downloaderType) {
	case "qbittorrent":
		d = NewQBittorrentDownloader()
	case "alist":
		// TODO: Alist 下载器待实现
		return nil, fmt.Errorf("alist 下载器暂未实现")
	default:
		return nil, fmt.Errorf("不支持的下载器类型: %s，支持的类型: qbittorrent, alist", downloaderType)
	}
	return d, nil
}
