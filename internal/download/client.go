package download

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"goto-bangumi/internal/apperrors"
	"goto-bangumi/internal/download/downloader"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/network"
)

// DownloadClient 下载客户端，负责限流和登录管理
type DownloadClient struct {
	Downloader     downloader.BaseDownloader
	limiter        *rate.Limiter
	SavePath       string
	downloaderType string

	// 登录控制
	loginDone  chan struct{} // 通知登录完成
	loginReq   chan struct{} // 是否需要登录
	loginError chan struct{} // 是否正在登录
}

// Client 为一个全局的下载客户端实例

var Client = &DownloadClient{}

func (c *DownloadClient) Init(config *model.DownloaderConfig) error {
	c.SavePath = config.SavePath

	downloaderType := strings.ToLower(config.Type)
	if c.downloaderType != downloaderType {
		c.downloaderType = downloaderType
		dl, err := downloader.NewDownloader(c.downloaderType)
		if err != nil {
			return fmt.Errorf("创建下载器失败: %w", err)
		}
		c.Downloader = dl

		// 从 downloader 获取 API interval（每个 downloader 自己定义）
		interval := dl.GetInterval()

		// 创建限流器：rate.Every 将间隔转换为速率
		c.limiter = rate.NewLimiter(rate.Every(time.Duration(interval)*time.Millisecond), 1)
	}
	// 初始化登录控制通道
	c.loginError = make(chan struct{})
	c.loginReq = make(chan struct{}, 1) // 缓冲区为1，避免重复登录请求
	c.loginReq <- struct{}{}            // 初始时请求登录
	c.loginDone = make(chan struct{})
	close(c.loginDone) // 初始时表示未登录
	return nil
}

func (c *DownloadClient) Check(ctx context.Context, hash string) (string, error) {
	if err := c.ensureLogin(ctx); err != nil {
		return "", fmt.Errorf("登录失败: %w", err)
	}

	if err := c.limiter.Wait(ctx); err != nil {
		return "", err
	}

	return c.Downloader.CheckHash(hash)
}

// RequestLogin 请求登录（非阻塞）
// 如果通道已满（已有登录请求），则直接返回，避免重复请求
func (c *DownloadClient) RequestLogin() {
	select {
	case c.loginReq <- struct{}{}:
		// 成功发送登录请求
	default:
		// 通道已满，已有登录请求，直接返回
	}
}

// Login 登录循环（后台运行）
// 从 loginReq 通道接收登录请求，处理登录逻辑
// 如果是网络错误，自动重试；如果是认证错误，等待外部重新请求
func (c *DownloadClient) Login(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("下载客户端登录协程退出")
			return
		case <-c.loginReq:
			// 执行登录
			c.loginDone = make(chan struct{})
			_, err := c.Downloader.Auth()

			if err != nil {
				if apperrors.IsNetworkError(err) {
					// 网络错误：等待30秒后自动重试
					time.Sleep(30 * time.Second)
					// 重新放入登录请求（阻塞式，确保重试）
					c.RequestLogin()
				} else if apperrors.IsDownloadAuthenticationError(err) || apperrors.IsDownloadForbiddenError(err) {
					slog.Error("下载客户端登录失败，认证错误，请检查配置", "error", err)
					close(c.loginError)
					// 这时不应该再自动重试了, resetLogin 也不该再触发登陆
					return
				}
				// 认证错误：不自动重试，等待外部调用 RequestLogin
			} else {
				// 登录成功
				// 通过关闭通道通知等待的请求
				close(c.loginDone)
				// 在登陆的过程中, 一些慢的请求会返回认证错误,然后触发重新登录
				select {
				case <-c.loginReq:
					// 清空成功
				default:
					// 通道为空，不等待
				}
			}
		}
	}
}

func (c *DownloadClient) ensureLogin(ctx context.Context) error {
	// 在logining 的时候, loginDone 会被重新赋值,这时会等待
	// loginDone 被关闭, 表明没有在登录
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.loginDone:
		return nil
	case <-c.loginError:
		return &apperrors.DownloadLoginError{Err: fmt.Errorf("下载协程已退出")}
	// 最多只等待10秒，避免长时间阻塞
	case <-time.After(10 * time.Second):
		return fmt.Errorf("等待登录超时")
	}
}

// Add 添加种子
func (c *DownloadClient) Add(ctx context.Context, url, savePath string) (string, error) {
	// 1. 确保已登录
	if err := c.ensureLogin(ctx); err != nil {
		return "", fmt.Errorf("登录失败: %w", err)
	}

	// 2. 限流
	if err := c.limiter.Wait(ctx); err != nil {
		return "", err
	}
	// 解析种子或磁力链接
	var torrentInfo *model.TorrentInfo
	var err error

	if !strings.HasPrefix(url, "magnet:") {
		// 下载种子文件
		networkClient := network.GetRequestClient()
		respBody, err := networkClient.Get(url)
		if err != nil {
			return "", fmt.Errorf("下载种子文件失败: %w", err)
		}

		// 解析种子文件
		torrentInfo, err = ParseTorrent(respBody)
		if err != nil {
			return "", fmt.Errorf("解析种子文件失败: %w", err)
		}
	} else {
		// 解析磁力链接
		torrentInfo, err = ParseTorrentURL(url)
		if err != nil {
			return "", fmt.Errorf("解析磁力链接失败: %w", err)
		}
	}

	// 3. 调用实际方法
	hash, err := c.Downloader.Add(torrentInfo, savePath)
	// 4. 如果是认证错误，重置登录状态
	if err != nil {
		if apperrors.IsDownloadAuthenticationError(err) {
			c.resetLogin()
		}
		return "", err
	}
	// TODO: 要对拿回来的 做一个 check, 会有很多情况, 比如有 v2但是qb 不认,所以这传回来的应该是个 list
	// 然后通过一些可能的 check , 来确认到底是哪一个
	// 5. check hash 拿到 真实的 hash
	// duid, err := c.Downloader.CheckHash(url)

	return hash, err
}

// Delete 删除种子
func (c *DownloadClient) Delete(ctx context.Context, hashes []string) error {
	if err := c.ensureLogin(ctx); err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}

	if err := c.limiter.Wait(ctx); err != nil {
		return err
	}

	_, err := c.Downloader.Delete(hashes)
	if err != nil && apperrors.IsDownloadAuthenticationError(err) {
		c.resetLogin()
	}

	return err
}

// Rename 重命名种子文件
func (c *DownloadClient) Rename(ctx context.Context, hash, oldPath, newPath string) error {
	if err := c.ensureLogin(ctx); err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}

	if err := c.limiter.Wait(ctx); err != nil {
		return err
	}

	_, err := c.Downloader.Rename(hash, oldPath, newPath)
	if err != nil && apperrors.IsDownloadAuthenticationError(err) {
		c.resetLogin()
	}

	return err
}

// Move 移动种子
func (c *DownloadClient) Move(ctx context.Context, hashes []string, location string) error {
	if err := c.ensureLogin(ctx); err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}

	if err := c.limiter.Wait(ctx); err != nil {
		return err
	}

	_, err := c.Downloader.Move(hashes, location)
	if err != nil && apperrors.IsDownloadAuthenticationError(err) {
		c.resetLogin()
	}

	return err
}

// GetTorrentFiles 获取种子文件列表
func (c *DownloadClient) GetTorrentFiles(ctx context.Context, hash string) ([]string, error) {
	if err := c.ensureLogin(ctx); err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}

	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	return c.Downloader.GetTorrentFiles(hash)
}

func (c *DownloadClient) GetTorrentInfo(ctx context.Context, hash string) (*model.TorrentDownloadInfo, error) {
	if err := c.ensureLogin(ctx); err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}

	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	return c.Downloader.GetTorrentInfo(hash)
}

// TorrentsInfo 获取种子信息列表
func (c *DownloadClient) TorrentsInfo(ctx context.Context, statusFilter, category string, tag *string, limit int) ([]map[string]interface{}, error) {
	if err := c.ensureLogin(ctx); err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}

	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	return c.Downloader.TorrentsInfo(statusFilter, category, tag, limit)
}

// resetLogin 重置登录状态（用于 403 错误后重新登录）
func (c *DownloadClient) resetLogin() {
	// 当能获得 c.loginDone 时，说明当前没有在登录
	// 这时请求登录
	// 否则说明已经在登录中，无需重复请求
	select {
	case <-c.loginDone:
		c.RequestLogin()
	default:
		// do nothing
	}
}
