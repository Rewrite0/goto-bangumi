// Package download 提供下载客户端，负责登录管理并转发下载器操作
package download

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/sync/singleflight"

	"goto-bangumi/internal/apperrors"
	"goto-bangumi/internal/download/downloader"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/network"
)

// DownloadClient 下载客户端，负责登录管理
type DownloadClient struct {
	Downloader     downloader.BaseDownloader
	SavePath       string
	downloaderType string

	// 登录控制
	logined    bool // 是否已登录
	LoginError bool // 登录错误通道
	loginGroup singleflight.Group
}

// NewDownloadClient 创建下载客户端实例
func NewDownloadClient() *DownloadClient {
	return &DownloadClient{}
}

func (c *DownloadClient) Init(config *model.DownloaderConfig) {
	c.SavePath = config.SavePath

	downloaderType := strings.ToLower(config.Type)
	if c.downloaderType != downloaderType {
		c.downloaderType = downloaderType
		dl := downloader.NewDownloader(c.downloaderType)
		c.Downloader = dl
	}
	c.Downloader.Init(config)
}

func (c *DownloadClient) Check(ctx context.Context, hash string) (string, error) {
	if err := c.EnsureLogin(ctx); err != nil {
		return "", fmt.Errorf("登录失败: %w", err)
	}

	return c.Downloader.CheckHash(ctx, hash)
}

// Login 登录（使用 singleflight 确保同一时间只有一个登录协程）
// 如果是网络错误，自动重试；如果是认证错误，等待外部重新请求
func (c *DownloadClient) Login(ctx context.Context) error {
	_, err, _ := c.loginGroup.Do("login", func() (any, error) {
		_, err := c.Downloader.Auth(ctx)
		if apperrors.IsDownloadAuthenticationError(err) || apperrors.IsDownloadForbiddenError(err) {
			c.LoginError = true
			return nil, apperrors.NewDownloadLoginError(
				fmt.Errorf("下载客户端认证失败，请检查配置: %w", err))
		}
		if err != nil {
			slog.Error("[download client]下载客户端登录失败，网络错误", "error", err)
			return nil, err
		}
		return nil, nil
	})
	if err == nil {
		// 登录成功，更新状态
		c.logined = true
	}
	return err
}

func (c *DownloadClient) EnsureLogin(ctx context.Context) error {
	if c.LoginError {
		return &apperrors.DownloadLoginError{Err: fmt.Errorf("下载器配置错误")}
	}
	if !c.logined {
		return c.Login(ctx)
	}
	return nil
}

// Add 添加种子
func (c *DownloadClient) Add(ctx context.Context, url, savePath string) ([]string, error) {
	// 1. 确保已登录
	if err := c.EnsureLogin(ctx); err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}

	// 解析种子或磁力链接
	var torrentInfo *model.TorrentInfo
	var err error

	if !strings.HasPrefix(url, "magnet:") {
		// 下载种子文件
		networkClient := network.GetRequestClient()
		respBody, err := networkClient.Get(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("[download client] 下载种子文件失败: %w", err)
		}

		// 解析种子文件
		torrentInfo, err = ParseTorrent(respBody)
		if err != nil {
			return nil, fmt.Errorf("[download client] 解析种子文件失败: %w", err)
		}
	} else {
		// 解析磁力链接
		torrentInfo, err = ParseTorrentURL(url)
		if err != nil {
			return nil, fmt.Errorf("[download client] 解析磁力链接失败: %w", err)
		}
	}

	// 3. 调用实际方法
	hashs, err := c.Downloader.Add(ctx, torrentInfo, savePath)
	// 4. 如果是认证错误，重置登录状态
	if err != nil {
		if apperrors.IsDownloadAuthenticationError(err) {
			c.logined = false
		}
		return nil, err
	}
	// TODO: 要对拿回来的 做一个 check, 会有很多情况, 比如有 v2但是qb 不认,所以这传回来的应该是个 list
	// 然后通过一些可能的 check , 来确认到底是哪一个
	// 5. check hash 拿到 真实的 hash
	// duid, err := c.Downloader.CheckHash(url)

	return hashs, err
}

// Delete 删除种子
func (c *DownloadClient) Delete(ctx context.Context, hashes []string) error {
	if err := c.EnsureLogin(ctx); err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}

	_, err := c.Downloader.Delete(ctx, hashes)
	if err != nil && apperrors.IsDownloadAuthenticationError(err) {
		c.logined = false
	}

	return err
}

// Rename 重命名种子文件
func (c *DownloadClient) Rename(ctx context.Context, hash, oldPath, newPath string) error {
	if err := c.EnsureLogin(ctx); err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}

	_, err := c.Downloader.Rename(ctx, hash, oldPath, newPath)
	if err != nil && apperrors.IsDownloadAuthenticationError(err) {
		c.logined = false
	}

	return err
}

// Move 移动种子
func (c *DownloadClient) Move(ctx context.Context, hashes []string, location string) error {
	if err := c.EnsureLogin(ctx); err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}

	_, err := c.Downloader.Move(ctx, hashes, location)
	if err != nil && apperrors.IsDownloadAuthenticationError(err) {
		c.logined = false
	}

	return err
}

// GetTorrentFiles 获取种子文件列表
func (c *DownloadClient) GetTorrentFiles(ctx context.Context, hash string) ([]string, error) {
	if err := c.EnsureLogin(ctx); err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}
	return c.Downloader.GetTorrentFiles(ctx, hash)
}

func (c *DownloadClient) GetTorrentInfo(ctx context.Context, hash string) (*model.TorrentDownloadInfo, error) {
	if err := c.EnsureLogin(ctx); err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}
	return c.Downloader.GetTorrentInfo(ctx, hash)
}

// TorrentsInfo 获取种子信息列表
func (c *DownloadClient) TorrentsInfo(ctx context.Context, statusFilter, category string, tag *string, limit int) ([]map[string]any, error) {
	if err := c.EnsureLogin(ctx); err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}

	return c.Downloader.TorrentsInfo(ctx, statusFilter, category, tag, limit)
}
