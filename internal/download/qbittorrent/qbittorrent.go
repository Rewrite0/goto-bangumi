package qbittorrent

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"goto-bangumi/internal/model"
)

//QBAPI qBittorrent API URL 定义
var QBAPI = map[string]string{
	"add":            "/api/v2/torrents/add",
	"addTags":        "/api/v2/torrents/addTags",
	"createCategory": "/api/v2/torrents/createCategory",
	"delete":         "/api/v2/torrents/delete",
	"getFiles":       "/api/v2/torrents/files",
	"info":           "/api/v2/torrents/info",
	"properties":     "/api/v2/torrents/properties",
	"login":          "/api/v2/auth/login",
	"logout":         "/api/v2/auth/logout",
	"renameFile":     "/api/v2/torrents/renameFile",
	"setCategory":    "/api/v2/torrents/setCategory",
	"setLocation":    "/api/v2/torrents/setLocation",
	"setPreferences": "/api/v2/app/setPreferences",
	"version":        "/api/v2/app/version",
}

// Downloader qBittorrent 下载器实现
type Downloader struct {
	client      *resty.Client
	config      *model.DownloaderConfig
	apiInterval float64
}

// NewDownloader 创建新的 qBittorrent 下载器
func NewDownloader() *Downloader {
	return &Downloader{
		apiInterval: 0.2,
	}
}

// Init 初始化下载器
func (d *Downloader) Init(config *model.DownloaderConfig) error {
	// 保存配置
	d.config = config

	// 创建 resty 客户端
	client := resty.New()
	client.SetBaseURL(d.config.Host)
	client.SetTimeout(30 * time.Second)

	// 设置默认请求头
	client.SetHeaders(map[string]string{
		"User-Agent": "goto-bangumi",
	})

	// 设置 TLS 配置
	// 如果不验证 SSL，则跳过证书验证
	if !d.config.Ssl {
		client.SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: true,
		})
	}

	d.client = client
	return nil
}

// Auth 认证登录
func (d *Downloader) Auth() (bool, error) {
	resp, err := d.client.R().
		SetFormData(map[string]string{
			"username": d.config.Username,
			"password": d.config.Password,
		}).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Post(QBAPI["login"])

	if err != nil {
		slog.Error("[qBittorrent] 连接到qBittorrent时出错，请检查您的主机配置", "host", d.config.Host, "error", err)
		return false, err
	}

	if resp.StatusCode() == 200 {
		if strings.TrimSpace(resp.String()) == "Ok." {
			slog.Debug("[qBittorrent] 登录成功")
			return true, nil
		}
		if strings.TrimSpace(resp.String()) == "Fails." {
			slog.Error("[qBittorrent] 登录失败，请检查用户名/密码", "username", d.config.Username)
			return false, fmt.Errorf("登录失败：用户名或密码错误")
		}
	}

	if resp.StatusCode() == 403 {
		slog.Error("[qBittorrent] 您的IP已被qBittorrent封禁，请解除封禁（或重启qBittorrent）后重试")
		return false, fmt.Errorf("IP被封禁")
	}

	return false, fmt.Errorf("登录失败：状态码 %d", resp.StatusCode())
}

// Logout 登出
func (d *Downloader) Logout() (bool, error) {
	resp, err := d.client.R().Post(QBAPI["logout"])
	if err != nil {
		slog.Error("[qBittorrent] 登出错误", "error", err)
		return false, err
	}

	if resp.StatusCode() == 200 {
		return true, nil
	}

	return false, fmt.Errorf("登出失败：状态码 %d", resp.StatusCode())
}

// CheckHost 检查主机连通性
func (d *Downloader) CheckHost() (bool, error) {
	slog.Debug("[qBittorrent] 检查主机", "host", d.config.Host)

	resp, err := d.client.R().Get(QBAPI["version"])
	if err != nil {
		slog.Error("[qBittorrent] 检查主机错误，请检查您的主机配置", "host", d.config.Host, "error", err)
		return false, err
	}

	// 状态码 200 或 403 都表示主机可访问
	if resp.StatusCode() == 200 || resp.StatusCode() == 403 {
		slog.Debug("[qBittorrent] 检查主机成功", "host", d.config.Host)
		return true, nil
	}

	return false, fmt.Errorf("主机不可访问：状态码 %d", resp.StatusCode())
}

// AddCategory 添加分类
func (d *Downloader) AddCategory(category string) (bool, error) {
	resp, err := d.client.R().
		SetFormData(map[string]string{
			"category": category,
		}).
		Post(QBAPI["createCategory"])

	if err != nil {
		d.handleException(err, resp, "add_category")
		return false, err
	}

	if resp.StatusCode() == 200 {
		return true, nil
	}

	return false, fmt.Errorf("添加分类失败：状态码 %d", resp.StatusCode())
}

// GetTorrentFiles 获取种子文件列表
func (d *Downloader) GetTorrentFiles(hash string) ([]string, error) {
	resp, err := d.client.R().
		SetQueryParam("hash", hash).
		Get(QBAPI["getFiles"])

	if err != nil {
		d.handleException(err, resp, "get_torrent_files")
		return nil, err
	}

	// 404 表示种子不存在
	if resp.StatusCode() == 404 {
		slog.Warn("[qBittorrent] 找不到种子", "hash", hash)
		return nil, nil
	}

	if strings.Contains(resp.String(), "Not Found") {
		slog.Warn("[qBittorrent] 找不到种子", "hash", hash)
		return nil, nil
	}

	if resp.StatusCode() == 200 {
		var files []map[string]interface{}
		if err := json.Unmarshal(resp.Body(), &files); err != nil {
			return nil, fmt.Errorf("解析文件列表失败: %w", err)
		}

		fileNames := make([]string, 0, len(files))
		for _, file := range files {
			if name, ok := file["name"].(string); ok {
				fileNames = append(fileNames, name)
			}
		}
		return fileNames, nil
	}

	return nil, fmt.Errorf("获取文件列表失败：状态码 %d", resp.StatusCode())
}

// TorrentInfo 获取单个种子详细信息
// 返回 (连接状态, 种子信息, error)
// 连接状态: true 表示连上了 client, false 表示没连上
// 种子信息: 成功时返回 TorrentDownloadInfo, 失败返回 nil
func (d *Downloader) TorrentInfo(hash string) (bool, *model.TorrentDownloadInfo, error) {
	resp, err := d.client.R().
		SetQueryParam("hash", hash).
		Get(QBAPI["properties"])

	if err != nil {
		// 连接错误
		slog.Error("[qBittorrent] torrent_info 连接错误", "error", err)
		return false, nil, err
	}

	// 404 表示种子不存在但连接正常
	if resp.StatusCode() == 404 {
		slog.Warn("[qBittorrent] 种子不存在", "hash", hash)
		return true, nil, nil
	}

	if resp.StatusCode() == 200 {
		var info map[string]interface{}
		if err := json.Unmarshal(resp.Body(), &info); err != nil {
			return true, nil, fmt.Errorf("解析种子信息失败: %w", err)
		}

		slog.Debug("[qBittorrent] 种子信息", "hash", hash, "eta", info["eta"], "save_path", info["save_path"], "completion_date", info["completion_date"])

		// 处理 completion_date
		completionDate := 0
		if cd, ok := info["completion_date"].(float64); ok && cd != -1 {
			completionDate = int(cd)
		}

		// 处理 eta
		eta := 0
		if e, ok := info["eta"].(float64); ok {
			eta = int(e)
		}

		result := &model.TorrentDownloadInfo{
			ETA:       &eta,
			SavePath:  info["save_path"].(string),
			Completed: completionDate,
		}

		return true, result, nil
	}

	d.handleException(err, resp, "torrent_info")
	return false, nil, fmt.Errorf("获取种子信息失败：状态码 %d", resp.StatusCode())
}

// TorrentsInfo 获取种子信息列表
func (d *Downloader) TorrentsInfo(statusFilter, category string, tag *string, limit int) ([]map[string]interface{}, error) {
	req := d.client.R().
		SetQueryParams(map[string]string{
			"filter":   statusFilter,
			"category": category,
			"sort":     "completion_on",
			"reverse":  "true",
		})

	if tag != nil {
		req.SetQueryParam("tag", *tag)
	}

	if limit > 0 {
		req.SetQueryParam("limit", fmt.Sprintf("%d", limit))
	}

	resp, err := req.Get(QBAPI["info"])

	if err != nil {
		d.handleException(err, resp, "torrents_info")
		return nil, err
	}

	if resp.StatusCode() == 200 {
		var torrents []map[string]interface{}
		if err := json.Unmarshal(resp.Body(), &torrents); err != nil {
			return nil, fmt.Errorf("解析种子列表失败: %w", err)
		}

		return torrents, nil
	}

	d.handleException(err, resp, "torrents_info")
	return nil, fmt.Errorf("获取种子列表失败：状态码 %d", resp.StatusCode())
}

// Add 添加种子
// TODO: 没必要这么复杂的 hash ,  加个tag 然后再拿一次就好了
func (d *Downloader) Add(torrentURL, savePath, category string) (*string, error) {
	var torrentHash string
	var torrentContent []byte

	slog.Debug("[qBittorrent] 开始获取种子内容", "url", torrentURL)

	// 判断是否为 magnet 链接
	if !strings.HasPrefix(torrentURL, "magnet:") {
		// 下载种子文件
		resp, err := d.client.R().Get(torrentURL)
		if err != nil || resp.StatusCode() != 200 {
			// 如果下载失败，尝试从 URL 提取 hash
			torrentHash = extractHash(torrentURL)
			if torrentHash == "" {
				slog.Error("[qBittorrent] 无法获取种子hash", "url", torrentURL)
				return nil, fmt.Errorf("无法获取种子hash")
			}
			slog.Warn("[qBittorrent] 无法获取种子内容，从URL提取hash", "url", torrentURL)
		} else {
			torrentContent = resp.Body()
			// TODO: 从种子文件中提取 hash
			torrentHash = extractHash(torrentURL)
			slog.Debug("[qBittorrent] 成功获取种子内容", "hash", torrentHash)
		}
	} else {
		// magnet 链接直接提取 hash
		torrentHash = extractHash(torrentURL)
		slog.Debug("[qBittorrent] 使用magnet链接", "url", torrentURL)
	}

	// 准备请求
	req := d.client.R().
		SetFormData(map[string]string{
			"savepath": savePath,
			"category": category,
			"paused":   "false",
			"autoTMM":  "false",
		})

	// 如果有种子文件内容，作为文件上传；否则使用 URL
	if len(torrentContent) > 0 {
		req.SetFileReader("torrents", "torrent.torrent", strings.NewReader(string(torrentContent)))
	} else {
		req.SetFormData(map[string]string{
			"urls":     torrentURL,
			"savepath": savePath,
			"category": category,
			"paused":   "false",
			"autoTMM":  "false",
		})
	}

	resp, err := req.Post(QBAPI["add"])

	if err != nil {
		d.handleException(err, resp, "add")
		return nil, err
	}

	if resp.StatusCode() == 200 {
		respText := strings.ToLower(resp.String())
		if strings.Contains(respText, "fail") {
			slog.Debug("[qBittorrent] 添加种子失败", "savepath", savePath, "response", respText)
			return nil, fmt.Errorf("添加种子失败: %s", respText)
		}

		// 返回 hash，只取前 40 个字符
		if len(torrentHash) > 40 {
			torrentHash = torrentHash[:40]
		}
		return &torrentHash, nil
	}

	d.handleException(err, resp, "add")
	return nil, fmt.Errorf("添加种子失败：状态码 %d", resp.StatusCode())
}

// Delete 删除种子
func (d *Downloader) Delete(hashes []string) (bool, error) {
	hashesStr := strings.Join(hashes, "|")

	resp, err := d.client.R().
		SetFormData(map[string]string{
			"hashes":      hashesStr,
			"deleteFiles": "true",
		}).
		Post(QBAPI["delete"])

	if err != nil {
		d.handleException(err, resp, "delete")
		return false, err
	}

	if resp.StatusCode() == 200 {
		return true, nil
	}

	d.handleException(err, resp, "delete")
	return false, fmt.Errorf("删除种子失败：状态码 %d", resp.StatusCode())
}

// Rename 重命名种子文件
func (d *Downloader) Rename(torrentHash, oldPath, newPath string) (bool, error) {
	resp, err := d.client.R().
		SetFormData(map[string]string{
			"hash":    torrentHash,
			"oldPath": oldPath,
			"newPath": newPath,
		}).
		Post(QBAPI["renameFile"])

	if err != nil {
		d.handleException(err, resp, "rename")
		return false, err
	}

	if resp.StatusCode() == 200 {
		return true, nil
	}

	// 409 表示文件已存在
	if resp.StatusCode() == 409 {
		slog.Error("[qBittorrent] 重命名错误，文件已存在", "oldPath", oldPath, "newPath", newPath)
		return false, fmt.Errorf("文件已存在")
	}

	d.handleException(err, resp, "rename")
	return false, fmt.Errorf("重命名失败：状态码 %d", resp.StatusCode())
}

// Move 移动种子到新位置
func (d *Downloader) Move(hashes []string, newLocation string) (bool, error) {
	hashesStr := strings.Join(hashes, "|")

	resp, err := d.client.R().
		SetFormData(map[string]string{
			"hashes":   hashesStr,
			"location": newLocation,
		}).
		Post(QBAPI["setLocation"])

	if err != nil {
		d.handleException(err, resp, "move")
		return false, err
	}

	if resp.StatusCode() == 200 {
		return true, nil
	}

	d.handleException(err, resp, "move")
	return false, fmt.Errorf("移动种子失败：状态码 %d", resp.StatusCode())
}

// SetCategory 设置种子分类
func (d *Downloader) SetCategory(hash, category string) (bool, error) {
	resp, err := d.client.R().
		SetFormData(map[string]string{
			"hashes":   hash,
			"category": category,
		}).
		Post(QBAPI["setCategory"])

	if err != nil {
		d.handleException(err, resp, "set_category")
		return false, err
	}

	if resp.StatusCode() == 200 {
		return true, nil
	}

	d.handleException(err, resp, "set_category")
	return false, fmt.Errorf("设置分类失败：状态码 %d", resp.StatusCode())
}

// AddTag 添加标签
func (d *Downloader) AddTag(hash, tag string) (bool, error) {
	resp, err := d.client.R().
		SetFormData(map[string]string{
			"hashes": hash,
			"tags":   tag,
		}).
		Post(QBAPI["addTags"])

	if err != nil {
		d.handleException(err, resp, "add_tag")
		return false, err
	}

	if resp.StatusCode() == 200 {
		return true, nil
	}

	d.handleException(err, resp, "add_tag")
	return false, fmt.Errorf("添加标签失败：状态码 %d", resp.StatusCode())
}

// SetPreferences 设置偏好设置
func (d *Downloader) SetPreferences(prefs map[string]interface{}) error {
	resp, err := d.client.R().
		SetBody(prefs).
		Post(QBAPI["setPreferences"])

	if err != nil {
		d.handleException(err, resp, "set_preferences")
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("设置偏好失败：状态码 %d", resp.StatusCode())
	}

	return nil
}

// handleException 统一处理异常
func (d *Downloader) handleException(err error, resp *resty.Response, functionName string) {
	if resp != nil && resp.StatusCode() == 403 {
		slog.Error("[qBittorrent] 需要先登录", "function", functionName)
		if resp != nil {
			slog.Debug("[qBittorrent] 错误响应", "function", functionName, "status", resp.StatusCode(), "body", resp.String()[:min(200, len(resp.String()))])
		}
	} else if err != nil {
		slog.Error("[qBittorrent] 错误", "function", functionName, "error", err)
	} else if resp != nil {
		slog.Error("[qBittorrent] HTTP错误", "function", functionName, "status", resp.StatusCode())
	}
}

// extractHash 从 URL 或 magnet 链接中提取 hash
func extractHash(url string) string {
	// 尝试从 magnet 链接提取
	if strings.HasPrefix(url, "magnet:") {
		re := regexp.MustCompile(`btih:([a-fA-F0-9]{40}|[A-Z2-7]{32})`)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return strings.ToLower(matches[1])
		}
	}

	// 尝试从 URL 路径提取 40 位十六进制 hash
	re := regexp.MustCompile(`[a-fA-F0-9]{40}`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 0 {
		return strings.ToLower(matches[0])
	}

	return ""
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
