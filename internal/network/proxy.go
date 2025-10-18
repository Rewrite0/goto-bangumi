package network

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"goto-bangumi/internal/model"
)

// export type ProxyType = ['http', 'https', 'socks5'];
// SetProxy creates a proxy URL from config
// 如果配置无效或不支持，则返回 nil
func SetProxy(config *model.ProxyConfig) *url.URL {
	supportProxyTypes := []string{"http", "socks5"}
	if config == nil || !config.Enable {
		return nil
	}

	// 只支持 http 和 socks5
	proxyType := config.Type
	if !slices.Contains(supportProxyTypes, proxyType) {
		slog.Warn("[Network] Unsupported proxy type", "type", proxyType)
		return nil
	}
	// Remove http:// prefix if present
	host := config.Host
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "socks5://")

	// Build auth string
	auth := ""
	if config.Username != "" {
		auth = fmt.Sprintf("%s:%s@", config.Username, config.Password)
	}

	// Build proxy URL
	proxyURL := fmt.Sprintf("%s://%s%s:%d", proxyType, auth, host, config.Port)
	urlParse, err := url.Parse(proxyURL)
	if err != nil {
		slog.Error("[Network] Invalid proxy URL", "error", err)
		return nil
	}
	return urlParse
}

// TestProxy tests if proxy connection works
func TestProxy(config *model.ProxyConfig) error {
	proxyURL := SetProxy(config)

	// Create client with proxy
	transport := &http.Transport{}
	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	// Test connection to baidu.com
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.baidu.com", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("[Network] Cannot connect to proxy, please check your proxy settings", "error", err)
		return fmt.Errorf("proxy connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("proxy test failed with status: %d", resp.StatusCode)
	}

	slog.Info("[Network] Proxy test successful")
	return nil
}

// CreateProxyTransport creates an HTTP transport with proxy support
func CreateProxyTransport(config *model.ProxyConfig) (*http.Transport, error) {
	proxyURL := SetProxy(config)

	transport := &http.Transport{
		MaxIdleConns:       100,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: false,
	}

	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
		slog.Info("[Network] Using proxy", "host", proxyURL.Host)
	}

	return transport, nil
}
