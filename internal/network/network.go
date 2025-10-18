// Package network provides HTTP client abstractions with proxy, caching, and retry support.
package network

import (
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
	Get(url string) ([]byte, error)
	Post(url string, contentType string, body io.Reader) ([]byte, error)
	Close() error
}

// ContentClient extends HTTPClient with content-specific methods
type ContentClient interface {
	HTTPClient
	GetJSON(url string) (map[string]interface{}, error)
	GetJSONTo(url string, v any) error
	GetRSS(url string) (*model.RSSXml, error)
	GetHTML(url string) (string, error)
	GetContent(url string) ([]byte, error)
	GetTorrents(url string) ([]model.Torrent, error)
	GetRSSTitle(url string) (string, error)
	PostData(url string, data map[string]string, files map[string][]byte) ([]byte, error)
}

// CacheManager provides caching functionality
type CacheManager interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte, ttl time.Duration)
	Delete(key string)
}

