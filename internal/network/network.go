// Package network provides HTTP client abstractions with proxy, caching, and retry support.
package network

import (
	"context"
	"io"
	"time"

	"goto-bangumi/internal/model"
)

// Default configuration values
const (
	DefaultTimeout      = 30 * time.Second
	DefaultRetries      = 3
	DefaultRetryDelay   = 5 * time.Second
	DefaultCacheTTL     = 60 * time.Second
	DefaultMaxCacheSize = 1000
	DefaultUserAgent    = "Mozilla/5.0"
)

// HTTPClient provides basic HTTP operations
type HTTPClient interface {
	Get(ctx context.Context, url string) ([]byte, error)
	Post(ctx context.Context, url string, contentType string, body io.Reader) ([]byte, error)
	CheckURL(ctx context.Context, url string) bool
	Close() error
}

// ContentClient extends HTTPClient with content-specific methods
type ContentClient interface {
	HTTPClient
	GetJSON(ctx context.Context, url string) (map[string]interface{}, error)
	GetRSS(ctx context.Context, url string) (*model.RSSXml, error)
	GetHTML(ctx context.Context, url string) (string, error)
	GetContent(ctx context.Context, url string) ([]byte, error)
	GetTorrents(ctx context.Context, url string, limit int) ([]model.Torrent, error)
	GetRSSTitle(ctx context.Context, url string) (string, error)
	PostData(ctx context.Context, url string, data map[string]string, files map[string][]byte) ([]byte, error)
}

// CacheManager provides caching functionality
type CacheManager interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
	ItemCount() int
}

