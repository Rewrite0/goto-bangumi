package network

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"goto-bangumi/internal/model"
)

// SetProxy creates a proxy URL from config
func SetProxy(config *model.ProxyConfig) (*url.URL, error) {
	if config == nil || !config.Enable {
		return nil, nil
	}

	// Remove http:// prefix if present
	host := config.Host
	if len(host) > 7 && host[:7] == "http://" {
		host = host[7:]
	}
	if len(host) > 8 && host[:8] == "https://" {
		host = host[8:]
	}

	// Build auth string
	auth := ""
	if config.Username != "" {
		auth = fmt.Sprintf("%s:%s@", config.Username, config.Password)
	}

	// Validate proxy type
	proxyType := config.Type
	if proxyType != "http" && proxyType != "https" {
		return nil, fmt.Errorf("unsupported proxy type: %s (only http/https supported)", proxyType)
	}

	// Build proxy URL
	proxyURL := fmt.Sprintf("%s://%s%s:%d", proxyType, auth, host, config.Port)
	return url.Parse(proxyURL)
}

// TestProxy tests if proxy connection works
func TestProxy(config *model.ProxyConfig) error {
	proxyURL, err := SetProxy(config)
	if err != nil {
		return fmt.Errorf("failed to create proxy URL: %w", err)
	}

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
	proxyURL, err := SetProxy(config)
	if err != nil {
		return nil, err
	}

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
