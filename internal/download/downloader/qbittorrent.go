package downloader

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"goto-bangumi/internal/apperrors"
	"goto-bangumi/internal/model"

	"github.com/go-resty/resty/v2"
)

// QBAPI qBittorrent API URL 定义
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

// QBittorrentDownloader qBittorrent 下载器实现
type QBittorrentDownloader struct {
	client      *resty.Client
	config      *model.DownloaderConfig
	ConfigName  string
	APIInterval int // API 调用间隔（毫秒），导出供 client 使用
}

// NewQBittorrentDownloader 创建新的 qBittorrent 下载器
func NewQBittorrentDownloader() *QBittorrentDownloader {
	return &QBittorrentDownloader{
		client:      resty.New(),
		APIInterval: 200, // 默认 200ms 间隔
		ConfigName:  "download",
	}
}

// Init 初始化下载器
func (d *QBittorrentDownloader) Init(config *model.DownloaderConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 保存配置
	d.config = config

	// 构建完整的 URL（添加协议前缀）
	baseURL := d.config.Host
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		if d.config.Ssl {
			baseURL = "https://" + baseURL
		} else {
			baseURL = "http://" + baseURL
		}
	}

	d.client.SetBaseURL(baseURL)
	d.client.SetTimeout(10 * time.Second)

	// 设置默认请求头
	d.client.SetHeaders(map[string]string{
		"User-Agent": "goto-bangumi",
	})

	// 设置 TLS 配置
	// 如果不验证 SSL，则跳过证书验证
	if !d.config.Ssl {
		d.client.SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: true,
		})
	}
	return nil
}

// Auth 认证登录
func (d *QBittorrentDownloader) Auth() (bool, error) {
	resp, err := d.client.R().
		SetFormData(map[string]string{
			"username": d.config.Username,
			"password": d.config.Password,
		}).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Post(QBAPI["login"])
	if err != nil {
		// 网络连接错误：无法连接到服务器（connection refused、timeout 等）
		slog.Error("[qBittorrent] 连接到qBittorrent时出错，请检查您的主机配置", "host", d.config.Host, "error", err)
		return false, &apperrors.NetworkError{
			Err:        fmt.Errorf("连接到qBittorrent时出错: %w", err),
			StatusCode: 0, // StatusCode = 0 表示网络连接错误
		}
	}

	if resp.StatusCode() == 200 {
		if strings.Contains(resp.String(), "Ok.") {
			return true, nil
		}
		if strings.Contains(resp.String(), "Fails") {
			// 认证错误：用户名或密码错误
			slog.Error("[qBittorrent] 登录失败，请检查用户名/密码", "username", d.config.Username)
			return false, &apperrors.NetworkError{Err: fmt.Errorf("用户名或密码错误"), StatusCode: 403}
		}
	}

	if resp.StatusCode() == 403 {
		slog.Error("[qBittorrent] 您的IP已被qBittorrent封禁，请解除封禁（或重启qBittorrent）后重试")
		return false, &apperrors.NetworkError{Err: fmt.Errorf("IP被封禁"), StatusCode: 403}
	}

	return false, &apperrors.NetworkError{Err: fmt.Errorf("登录失败：状态码 %d", resp.StatusCode()), StatusCode: resp.StatusCode()}
}

// Logout 登出
func (d *QBittorrentDownloader) Logout() (bool, error) {
	resp, err := d.client.R().Post(QBAPI["logout"])
	if err != nil {
		slog.Error("[qBittorrent] 登出错误", "error", err)
		return false, &apperrors.NetworkError{Err: fmt.Errorf("登出错误: %w", err), StatusCode: 0}
	}

	if resp.StatusCode() == 200 {
		return true, nil
	}

	return false, &apperrors.NetworkError{Err: fmt.Errorf("登出失败：状态码 %d", resp.StatusCode()), StatusCode: resp.StatusCode()}
}

// CheckHost 检查主机连通性
// func (d *QBittorrentDownloader) CheckHost() (bool, error) {
// 	slog.Debug("[qBittorrent] 检查主机", "host", d.config.Host)
//
// 	// resp, err := d.client.R().Get(QBAPI["version"])
// 	// if err != nil {
// 	// 	slog.Error("[qBittorrent] 检查主机错误，请检查您的主机配置", "host", d.config.Host, "error", err)
// 	// 	return false, &apperrors.NetworkError{Err: fmt.Errorf("检查主机错误: %w", err), StatusCode: 0}
// 	// }
// 	//
// 	// // 状态码 200 或 403 都表示主机可访问
// 	// if resp.StatusCode() == 200 || resp.StatusCode() == 403 {
// 	// 	slog.Debug("[qBittorrent] 检查主机成功", "host", d.config.Host)
// 	// 	return true, nil
// 	// }
// 	//
// 	// return false, &apperrors.NetworkError{Err: fmt.Errorf("主机不可访问：状态码 %d", resp.StatusCode()), StatusCode: resp.StatusCode()}
// }

// AddCategory 添加分类
func (d *QBittorrentDownloader) AddCategory(category string) (bool, error) {
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
func (d *QBittorrentDownloader) GetTorrentFiles(hash string) ([]string, error) {
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

	fmt.Print("qb get files resp = ", resp.String())
	if resp.StatusCode() == 200 {
		var files []model.QBTorrentFile
		if err := json.Unmarshal(resp.Body(), &files); err != nil {
			return nil, fmt.Errorf("解析文件列表失败: %w", err)
		}

		fileNames := make([]string, 0, len(files))
		for _, file := range files {
			fileNames = append(fileNames, file.Name)
		}
		return fileNames, nil
	}

	return nil, fmt.Errorf("获取文件列表失败：状态码 %d", resp.StatusCode())
}

// TorrentInfo 获取单个种子详细信息
// 返回 (连接状态, 种子信息, error)
// 连接状态: true 表示连上了 client, false 表示没连上
// 种子信息: 成功时返回 TorrentDownloadInfo, 失败返回 nil
func (d *QBittorrentDownloader) GetTorrentInfo(hash string) (bool, *model.TorrentDownloadInfo, error) {
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
		var props model.QBTorrentProperties
		if err := json.Unmarshal(resp.Body(), &props); err != nil {
			return true, nil, fmt.Errorf("解析种子信息失败: %w", err)
		}

		slog.Debug("[qBittorrent] 种子信息", "hash", hash, "eta", props.Eta, "save_path", props.SavePath, "completion_date", props.CompletionDate)

		// 处理 completion_date
		completionDate := 0
		if props.CompletionDate != -1 {
			completionDate = int(props.CompletionDate)
		}

		result := &model.TorrentDownloadInfo{
			ETA:       int(props.Eta),
			SavePath:  props.SavePath,
			Completed: completionDate,
		}

		return true, result, nil
	}

	d.handleException(err, resp, "torrent_info")
	return false, nil, fmt.Errorf("获取种子信息失败：状态码 %d", resp.StatusCode())
}

// TorrentsInfo 获取种子信息列表
func (d *QBittorrentDownloader) TorrentsInfo(statusFilter, category string, tag *string, limit int) ([]map[string]interface{}, error) {
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
		var torrents []model.QBTorrentInfo
		if err := json.Unmarshal(resp.Body(), &torrents); err != nil {
			return nil, fmt.Errorf("解析种子列表失败: %w", err)
		}

		// 转换回 map[string]interface{} 以保持返回值兼容性
		result := make([]map[string]interface{}, 0, len(torrents))
		for _, t := range torrents {
			data, _ := json.Marshal(t)
			var m map[string]interface{}
			json.Unmarshal(data, &m)
			result = append(result, m)
		}

		return result, nil
	}

	d.handleException(err, resp, "torrents_info")
	return nil, fmt.Errorf("获取种子列表失败：状态码 %d", resp.StatusCode())
}

func (d *QBittorrentDownloader) CheckHash(hash string) (string, error) {
	return hash, nil
}

// Add 添加种子
// TODO: 没必要这么复杂的 hash ,  加个tag 然后再拿一次就好了
func (d *QBittorrentDownloader) Add(torrentInfo *model.TorrentInfo, savePath string) (string, error) {
	// 准备基础表单数据
	data := make(map[string]string)
	data["savepath"] = savePath
	data["category"] = "GotoBangumi"
	data["paused"] = "false"
	data["autoTMM"] = "false"

	var resp *resty.Response
	var err error

	// 如果有种子文件内容，作为文件上传；否则使用磁力链接
	if len(torrentInfo.File) > 0 {
		// 上传种子文件（二进制内容）
		resp, err = d.client.R().
			SetFormData(data).
			SetFileReader("torrents", "torrent.torrent", bytes.NewReader(torrentInfo.File)).
			Post(QBAPI["add"])
	} else {
		// 使用磁力链接
		data["urls"] = torrentInfo.MagnetURI
		resp, err = d.client.R().
			SetFormData(data).
			Post(QBAPI["add"])
	}

	if err != nil {
		return "", &apperrors.NetworkError{Err: fmt.Errorf("添加种子失败: %w", err), StatusCode: 0}
	}
	if resp.StatusCode() == 403 {
		slog.Error("[qBittorrent] 需要先登录", "function", "add")
		return "", &apperrors.NetworkError{Err: fmt.Errorf("需要先登录"), StatusCode: 403}
	}
	if resp.StatusCode() == 200 {
		respText := strings.ToLower(resp.String())
		if strings.Contains(respText, "fail") {
			slog.Debug("[qBittorrent] 添加种子失败, 种子重复", "savepath", savePath, "response", respText)
			return "", fmt.Errorf("添加种子失败: %s", respText)
		}
		return torrentInfo.InfoHashV1, nil
	}
	return "", fmt.Errorf("添加种子失败：状态码 %d", resp.StatusCode())
}

// Delete 删除种子
func (d *QBittorrentDownloader) Delete(hashes []string) (bool, error) {
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
func (d *QBittorrentDownloader) Rename(torrentHash, oldPath, newPath string) (bool, error) {
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
func (d *QBittorrentDownloader) Move(hashes []string, newLocation string) (bool, error) {
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
func (d *QBittorrentDownloader) SetCategory(hash, category string) (bool, error) {
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
func (d *QBittorrentDownloader) AddTag(hash, tag string) (bool, error) {
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
func (d *QBittorrentDownloader) SetPreferences(prefs map[string]interface{}) error {
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
func (d *QBittorrentDownloader) handleException(err error, resp *resty.Response, functionName string) {
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

func (d *QBittorrentDownloader) GetInterval() int {
	return d.APIInterval
}
