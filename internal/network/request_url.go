package network

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/singleflight"
	"goto-bangumi/internal/model"
)

// 包级变量，存储代理配置、缓存管理器和请求去重
var (
	defaultProxyConfig *model.ProxyConfig
	globalCache        CacheManager
	requestGroup       singleflight.Group
)

func init() {
	// 初始化全局缓存管理器（500 个缓存项，60 秒 TTL）
	globalCache = NewMemoryCacheManager(500, 60*time.Second)
	// 初始化默认代理配置为空
	defaultProxyConfig = &model.ProxyConfig{}
}

// Init 初始化 network 包的代理配置
func Init(config *model.ProxyConfig) {
	if config != nil {
		defaultProxyConfig = config
		slog.Info("[Network] Network package initialized", "proxy_enabled", config.Enable)
	}
}

// RequestClient provides HTTP request functionality with retry and proxy support using resty
type RequestClient struct {
	client *resty.Client
}

// NewRequestClient creates a new RequestURL instance with resty
func NewRequestClient() (*RequestClient, error) {

	// 如果没有init config，则使用包级的 defaultProxyConfig
	proxyConfig := defaultProxyConfig

	client := resty.New()

	// Set timeout
	client.SetTimeout(DefaultTimeout)

	// Set retry configuration
	client.SetRetryCount(DefaultRetries).
		SetRetryWaitTime(DefaultRetryDelay).
		SetRetryMaxWaitTime(DefaultRetryDelay * 2)

	// Set default headers
	client.SetHeaders(map[string]string{
		"User-Agent": DefaultUserAgent,
		"Accept":     "application/xml",
	})

	// Configure proxy if enabled
	if proxyConfig.Enable {
		proxyURL, err := SetProxy(proxyConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to set proxy: %w", err)
		}
		if proxyURL != nil {
			client.SetProxy(proxyURL.String())
			slog.Info("[Network] Using proxy", "host", proxyURL.Host)
		}
	}

	// Add retry condition: retry on 5xx errors and network errors
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		if err != nil {
			slog.Warn("[Network] Retrying due to error", "error", err)
			return true // Retry on network errors
		}
		// Retry on 5xx server errors
		if r.StatusCode() >= 500 {
			slog.Warn("[Network] Retrying due to server error",
				"url", r.Request.URL,
				"status", r.StatusCode())
			return true
		}
		return false
	})

	return &RequestClient{
		client: client,
	}, nil
}

// Get performs HTTP GET request with cache support and request deduplication
func (r *RequestClient) Get(url string) ([]byte, error) {
	// 1. 快速路径：检查缓存
	if data, found := globalCache.Get(url); found {
		slog.Debug("[Network] Cache hit", "url", url)
		return data, nil
	}

	// 2. 使用 singleflight 防止并发重复请求
	v, err, shared := requestGroup.Do(url, func() (interface{}, error) {
		// 2.2 执行实际 HTTP 请求
		fmt.Println("Fetching URL:", url)
		slog.Debug("[Network] Executing HTTP request", "url", url)
		resp, err := r.client.R().Get(url)
		if err != nil {
			return nil, fmt.Errorf("GET request failed: %w", err)
		}

		if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode(), resp.Status())
		}

		body := resp.Body()

		// 2.3 写入缓存
		globalCache.Set(url, body, DefaultCacheTTL)
		slog.Debug("[Network] Cached response", "url", url, "size", len(body))

		return body, nil
	})

	if err != nil {
		return nil, err
	}

	if shared {
		slog.Debug("[Network] Request shared via singleflight", "url", url)
	}

	return v.([]byte), nil
}

// Post performs HTTP POST request
func (r *RequestClient) Post(url string, contentType string, body io.Reader) ([]byte, error) {
	resp, err := r.client.R().
		SetHeader("Content-Type", contentType).
		SetBody(body).
		Post(url)

	if err != nil {
		return nil, fmt.Errorf("POST request failed: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("POST request failed with status: %d", resp.StatusCode())
	}

	return resp.Body(), nil
}

// CheckURL checks if a URL is accessible
func (r *RequestClient) CheckURL(urlStr string) bool {
	// Add http:// prefix if missing
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "http://" + urlStr
	}

	resp, err := r.client.R().
		Get(urlStr)

	if err != nil {
		slog.Debug("[Network] Cannot connect to URL", "url", urlStr, "error", err)
		return false
	}

	return resp.StatusCode() >= 200 && resp.StatusCode() < 400
}

// SetHeader sets a custom header
func (r *RequestClient) SetHeader(key, value string) {
	r.client.SetHeader(key, value)
}

// SetRetry sets the retry count
func (r *RequestClient) SetRetry(retry int) {
	r.client.SetRetryCount(retry)
}

// Close closes the HTTP client
func (r *RequestClient) Close() error {
	// Resty doesn't require explicit cleanup, but we can close the underlying HTTP client
	r.client.GetClient().CloseIdleConnections()
	return nil
}

// GetJSON performs GET request and returns parsed JSON as map
func (r *RequestClient) GetJSON(url string) (map[string]interface{}, error) {
	resp, err := r.client.R().
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("GET JSON request failed: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode(), resp.Status())
	}

	// Parse JSON
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

// GetRSS fetches and parses RSS feed, returns the parsed RSS object
func (r *RequestClient) GetRSS(url string) (*model.RSSXml, error) {
	resp, err := r.client.R().
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("GET RSS request failed: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode(), resp.Status())
	}

	// Parse XML
	var rss model.RSSXml
	if err := xml.Unmarshal(resp.Body(), &rss); err != nil {
		return nil, fmt.Errorf("failed to parse RSS XML: %w", err)
	}


	return &rss, nil
}

// GetHTML performs GET request and returns HTML as string
func (r *RequestClient) GetHTML(url string) (string, error) {
	resp, err := r.client.R().
		Get(url)

	if err != nil {
		return "", fmt.Errorf("GET HTML request failed: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode(), resp.Status())
	}

	return resp.String(), nil
}

// GetContent 和 Get 方法相同，后面会加缓存, Get 则是一个简单的 GET 请求
func (r *RequestClient) GetContent(url string) ([]byte, error) {
	return r.Get(url)
}

// GetTorrents fetches and parses RSS feed to extract torrents
func (r *RequestClient) GetTorrents(url string) ([]model.Torrent, error) {
	rss, err := r.GetRSS(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get RSS feed: %w", err)
	}

	torrents := make([]model.Torrent, 0, len(rss.Torrents))
	for _, item := range rss.Torrents {
		torrent := model.Torrent{
			Name:     item.Name,
			Homepage: item.Enclosure.URL,
			RssLink: url,
		}

		if item.Enclosure.URL != "" {
			torrent.URL = item.Enclosure.URL
			torrent.Homepage = item.Link
		} else {
			torrent.URL = item.Link
		}

		torrents = append(torrents, torrent)
	}

	return torrents, nil
}

// GetRSSTitle fetches RSS feed and returns the channel title
func (r *RequestClient) GetRSSTitle(url string) (string, error) {
	rss, err := r.GetRSS(url)
	if err != nil {
		return "", fmt.Errorf("failed to get RSS feed: %w", err)
	}

	return rss.Title, nil
}

// PostData sends form data and files via POST request
func (r *RequestClient) PostData(url string, data map[string]string, files map[string][]byte) ([]byte, error) {
	req := r.client.R()

	// Set form data
	if data != nil {
		req.SetFormData(data)
	}

	// Set files
	if files != nil {
		for name, content := range files {
			req.SetFileReader(name, name, strings.NewReader(string(content)))
		}
	}

	resp, err := req.Post(url)
	if err != nil {
		return nil, fmt.Errorf("POST data request failed: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("POST request failed with status: %d", resp.StatusCode())
	}

	return resp.Body(), nil
}
