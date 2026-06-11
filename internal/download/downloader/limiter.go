package downloader

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type apiLimiter struct {
	mu      sync.RWMutex
	limiter *rate.Limiter
}

func newAPILimiter(interval time.Duration) *apiLimiter {
	l := &apiLimiter{}
	l.SetInterval(interval)
	return l
}

func newAPILimiterFromQPS(qps float64) *apiLimiter {
	l := &apiLimiter{}
	l.SetQPS(qps)
	return l
}

func (l *apiLimiter) SetInterval(interval time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if interval <= 0 {
		l.limiter = nil
		return
	}
	l.limiter = rate.NewLimiter(rate.Every(interval), 1)
}

func (l *apiLimiter) SetQPS(qps float64) {
	if qps <= 0 {
		l.SetInterval(0)
		return
	}
	l.SetInterval(time.Duration(float64(time.Second) / qps))
}

func (l *apiLimiter) Wait(ctx context.Context, name string) error {
	if l == nil {
		return nil
	}

	l.mu.RLock()
	limiter := l.limiter
	l.mu.RUnlock()
	if limiter == nil {
		return nil
	}

	if err := limiter.Wait(ctx); err != nil {
		slog.Debug("["+name+"] request interrupted by user", "error", err)
		return err
	}
	return nil
}
