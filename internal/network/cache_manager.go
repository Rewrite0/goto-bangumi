package network

import (
	"log/slog"
	"sync"
	"time"
)

// cacheItem represents a cached entry with its data and timestamp
type cacheItem struct {
	data      []byte
	timestamp time.Time
}

// MemoryCacheManager implements CacheManager with in-memory storage
type MemoryCacheManager struct {
	cache    map[string]*cacheItem
	mu       sync.RWMutex
	maxSize  int
	cacheTTL time.Duration
}

// NewMemoryCacheManager creates a new in-memory cache manager
func NewMemoryCacheManager(maxSize int, ttl time.Duration) *MemoryCacheManager {
	return &MemoryCacheManager{
		cache:    make(map[string]*cacheItem),
		maxSize:  maxSize,
		cacheTTL: ttl,
	}
}

// Get retrieves a value from cache if it exists and hasn't expired (lazy deletion)
func (m *MemoryCacheManager) Get(key string) ([]byte, bool) {
	m.mu.RLock()
	item, exists := m.cache[key]
	m.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Check if expired (lazy deletion)
	if time.Since(item.timestamp) > m.cacheTTL {
		// Upgrade to write lock and delete expired item
		m.mu.Lock()
		// Double-check the item still exists and is still expired
		if item, exists := m.cache[key]; exists && time.Since(item.timestamp) > m.cacheTTL {
			delete(m.cache, key)
			slog.Debug("[Cache] Lazy deleted expired entry", "key", key)
		}
		m.mu.Unlock()
		return nil, false
	}

	return item.data, true
}

// Set stores a value in cache with the configured TTL
func (m *MemoryCacheManager) Set(key string, value []byte, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if cache is full
	if len(m.cache) >= m.maxSize {
		// Batch cleanup expired entries first
		currentTime := time.Now()
		cleanedCount := 0
		for k, v := range m.cache {
			if currentTime.Sub(v.timestamp) > m.cacheTTL {
				delete(m.cache, k)
				cleanedCount++
			}
		}
		if cleanedCount > 0 {
			slog.Debug("[Cache] Batch cleaned expired entries on full", "count", cleanedCount, "remaining", len(m.cache))
		}

		// If still full after cleanup, delete the oldest entry
		if len(m.cache) >= m.maxSize {
			var oldestKey string
			var oldestTime time.Time
			first := true

			for k, v := range m.cache {
				if first || v.timestamp.Before(oldestTime) {
					oldestKey = k
					oldestTime = v.timestamp
					first = false
				}
			}

			if oldestKey != "" {
				delete(m.cache, oldestKey)
				slog.Debug("[Cache] Evicted oldest entry due to size limit", "key", oldestKey)
			}
		}
	}

	// Store the new item
	m.cache[key] = &cacheItem{
		data:      value,
		timestamp: time.Now(),
	}
}

// Delete removes a specific key from cache
func (m *MemoryCacheManager) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.cache, key)
}
