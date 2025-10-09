package network

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/go-resty/resty/v2"
	"goto-bangumi/internal/model"
)

// RequestURL provides HTTP request functionality with retry and proxy support using resty
type RequestURL struct {
	client *resty.Client
}

// NewRequestURL creates a new RequestURL instance with resty
func NewRequestURL(config *model.ProxyConfig) (*RequestURL, error) {
	if config == nil {
		config = &model.ProxyConfig{}
	}

	// Create resty client
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
	if config.Enable {
		proxyURL, err := SetProxy(config)
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

	return &RequestURL{
		client: client,
	}, nil
}

// Get performs HTTP GET request
func (r *RequestURL) Get(ctx context.Context, url string) ([]byte, error) {
	resp, err := r.client.R().
		SetContext(ctx).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("GET request failed: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode(), resp.Status())
	}

	return resp.Body(), nil
}

// Post performs HTTP POST request
func (r *RequestURL) Post(ctx context.Context, url string, contentType string, body io.Reader) ([]byte, error) {
	resp, err := r.client.R().
		SetContext(ctx).
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
func (r *RequestURL) CheckURL(ctx context.Context, urlStr string) bool {
	// Add http:// prefix if missing
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "http://" + urlStr
	}

	resp, err := r.client.R().
		SetContext(ctx).
		Get(urlStr)

	if err != nil {
		slog.Debug("[Network] Cannot connect to URL", "url", urlStr, "error", err)
		return false
	}

	return resp.StatusCode() >= 200 && resp.StatusCode() < 400
}

// SetHeader sets a custom header
func (r *RequestURL) SetHeader(key, value string) {
	r.client.SetHeader(key, value)
}

// SetRetry sets the retry count
func (r *RequestURL) SetRetry(retry int) {
	r.client.SetRetryCount(retry)
}

// Close closes the HTTP client
func (r *RequestURL) Close() error {
	// Resty doesn't require explicit cleanup, but we can close the underlying HTTP client
	r.client.GetClient().CloseIdleConnections()
	return nil
}

// GetJSON performs GET request and returns parsed JSON as map
func (r *RequestURL) GetJSON(ctx context.Context, url string) (map[string]interface{}, error) {
	resp, err := r.client.R().
		SetContext(ctx).
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
func (r *RequestURL) GetRSS(ctx context.Context, url string) (*model.RSSXml, error) {
	resp, err := r.client.R().
		SetContext(ctx).
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
func (r *RequestURL) GetHTML(ctx context.Context, url string) (string, error) {
	resp, err := r.client.R().
		SetContext(ctx).
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
func (r *RequestURL) GetContent(ctx context.Context, url string) ([]byte, error) {
	return r.Get(ctx, url)
}

// GetTorrents fetches and parses RSS feed to extract torrents
func (r *RequestURL) GetTorrents(ctx context.Context, url string) ([]model.Torrent, error) {
	rss, err := r.GetRSS(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get RSS feed: %w", err)
	}

	torrents := make([]model.Torrent, 0, len(rss.Torrents))
	for _, item := range rss.Torrents {
		torrent := model.Torrent{
			Name:     item.Name,
			Homepage: item.Homepage.URL,
			RssLink: url,
		}

		if item.Homepage.URL != "" {
			torrent.URL = item.Homepage.URL
			torrent.Homepage = item.Link
		} else {
			torrent.URL = item.Link
		}

		torrents = append(torrents, torrent)
	}

	return torrents, nil
}

// GetRSSTitle fetches RSS feed and returns the channel title
func (r *RequestURL) GetRSSTitle(ctx context.Context, url string) (string, error) {
	rss, err := r.GetRSS(ctx, url)
	if err != nil {
		return "", fmt.Errorf("failed to get RSS feed: %w", err)
	}

	return rss.Title, nil
}

// PostData sends form data and files via POST request
func (r *RequestURL) PostData(ctx context.Context, url string, data map[string]string, files map[string][]byte) ([]byte, error) {
	req := r.client.R().SetContext(ctx)

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
