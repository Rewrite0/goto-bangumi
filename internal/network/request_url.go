package network

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"goto-bangumi/internal/apperrors"
	"goto-bangumi/internal/model"

	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/singleflight"
)

// 包级变量，存储代理配置、缓存管理器和请求去重
var (
	defaultProxyConfig *model.ProxyConfig
	globalCache        CacheManager
	requestGroup       singleflight.Group
	defaultClient      *RequestClient
)

func init() {
	// 初始化全局缓存管理器（500 个缓存项，60 秒 TTL）
	globalCache = NewMemoryCacheManager(500, 60*time.Second)
	// 初始化默认代理配置为空
	defaultProxyConfig = model.NewProxyConfig()
	defaultClient = newRequestClient()
}

// Init 初始化 network 包的代理配置
func Init(config *model.ProxyConfig) {
	if config != nil {
		defaultProxyConfig = config
		slog.Info("[Network] Network package initialized", "proxy_enabled", config.Enable)
		defaultClient = newRequestClient()
	}
}

// GetRequestClient 返回全局共享的 RequestClient 实例
// 所有 HTTP 请求应使用此实例以共享连接池和缓存
func GetRequestClient() *RequestClient {
	return defaultClient
}

// RequestClient provides HTTP request functionality with retry and proxy support using resty
type RequestClient struct {
	client *resty.Client
}

// NewRequestClient creates a new RequestURL instance with resty
func newRequestClient() *RequestClient {
	// 如果没有init config，则使用包级的 defaultProxyConfig
	proxyConfig := defaultProxyConfig

	// 创建自定义 transport，提高同 host 的连接数
	transport := &http.Transport{
		MaxIdleConns:        100,              // 最大空闲连接数
		MaxIdleConnsPerHost: 100,              // 每个 host 的最大空闲连接数（默认2）
		MaxConnsPerHost:     10,               // 每个 host 的最大连接数（默认0无限制）
		IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
	}

	client := resty.New()

	// 设置自定义 transport
	client.SetTransport(transport)

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
		proxyURL := SetProxy(proxyConfig)
		// 对应 config 里面的 Enable = false 的情况
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
	}
}

// Get performs HTTP GET request with cache support and request deduplication
func (r *RequestClient) Get(url string) ([]byte, error) {
	// 1. 快速路径：检查缓存
	if data, found := globalCache.Get(url); found {
		slog.Debug("[Network] Cache hit", "url", url)
		return data, nil
	}

	// 2. 使用 singleflight 防止并发重复请求
	v, err, shared := requestGroup.Do(url, func() (any, error) {
		// 2.2 执行实际 HTTP 请求
		fmt.Println("Fetching URL:", url)
		slog.Debug("[Network] Executing HTTP request", "url", url)
		resp, err := r.client.R().Get(url)
		if err != nil {
			return nil, &apperrors.NetworkError{Err: fmt.Errorf("GET request failed: %w", err), StatusCode: 0}
		}

		if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
			return nil, &apperrors.NetworkError{
				Err:        fmt.Errorf("HTTP %d: %s", resp.StatusCode(), resp.Status()),
				StatusCode: resp.StatusCode(),
			}
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
		return nil, &apperrors.NetworkError{Err: fmt.Errorf("POST request failed: %w", err), StatusCode: 0}
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return nil, &apperrors.NetworkError{
			Err:        fmt.Errorf("POST request failed with status: %d", resp.StatusCode()),
			StatusCode: resp.StatusCode(),
		}
	}

	return resp.Body(), nil
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
func (r *RequestClient) GetJSON(url string) (map[string]any, error) {
	resp, err := r.Get(url)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var result map[string]any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, &apperrors.ParseError{Err: fmt.Errorf("failed to parse JSON: %w", err)}
	}

	return result, nil
}

// GetJSONTo performs GET request and unmarshals JSON response into the provided value
// Usage: var result model.SearchResult; err := client.GetJSONTo(url, &result)
func (r *RequestClient) GetJSONTo(url string, v any) error {
	resp, err := r.Get(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp, v); err != nil {
		return &apperrors.ParseError{Err: fmt.Errorf("failed to parse JSON: %w", err)}
	}

	return nil
}

// GetRSS fetches and parses RSS feed, returns the parsed RSS object
func (r *RequestClient) GetRSS(url string) (*model.RSSXml, error) {
	// 需要能判断出来是网络不好还是空的 xml
	// 空的 https://mikanani.me/RSS/Search?searchstr=ANININI
	resp, err := r.Get(url) // 这里是网络问题
	if err != nil {
		return nil, err
	}
	// Parse XML
	var rss model.RSSXml
	if err := xml.Unmarshal(resp, &rss); err != nil {
		return nil, &apperrors.ParseError{Err: fmt.Errorf("failed to parse RSS XML: %w", err)}
	}
	return &rss, nil
}

// GetTorrents fetches and parses RSS feed to extract torrents
// 返回错误主是是区分是网络请求错误还是确实没有种子
func (r *RequestClient) GetTorrents(url string) ([]*model.Torrent, error) {
	rss, err := r.GetRSS(url)
	if err != nil {
		return nil, err
	}

	torrents := make([]*model.Torrent, 0, len(rss.Torrents))
	for _, item := range rss.Torrents {
		// 移除名称中的换行符和多余空格
		item.Name = processTitle(item.Name)
		// 创建 Torrent 对象
		torrent := &model.Torrent{
			Name:     item.Name,
			Homepage: item.Enclosure.URL,
			URL:      url,
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
	// GetRSS 已经包装了 NetworkError 或 ParseError，直接传递
	if err != nil {
		return "", err
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
		return nil, &apperrors.NetworkError{Err: fmt.Errorf("POST data request failed: %w", err), StatusCode: 0}
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return nil, &apperrors.NetworkError{
			Err:        fmt.Errorf("POST request failed with status: %d", resp.StatusCode()),
			StatusCode: resp.StatusCode(),
		}
	}

	return resp.Body(), nil
}

func processTitle(title string) string {
	// title 里面可能有"\n"
	title = strings.ReplaceAll(title, "\n", "")
	// 如果以【开头
	if strings.HasPrefix(title, "【") {
		title = strings.ReplaceAll(title, "【", "[")
		title = strings.ReplaceAll(title, "】", "]")
	}
	title = strings.TrimSpace(title)
	return title
}

// SetTestCache 用于测试时向全局缓存添加模拟数据
// 这个函数只应该在测试代码中使用
func SetTestCache(url string, data []byte) {
	globalCache.Set(url, data, DefaultCacheTTL)
}

// ClearTestCache 用于测试时清空指定 URL 的缓存
// 这个函数只应该在测试代码中使用
func ClearTestCache(url string) {
	globalCache.Delete(url)
}
