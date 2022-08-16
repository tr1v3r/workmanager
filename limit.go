package workmanager

import (
	"context"
	"runtime"
	"sync"

	"golang.org/x/time/rate"
)

func NewLimitManager(context.Context) *limitManager {
	return &limitManager{
		limiterMap:     make(map[WorkStep]*rate.Limiter),
		defaultLimiter: rate.NewLimiter(rate.Limit(runtime.NumCPU()*100), 100),
	}
}

type limitManager struct {
	mu             sync.RWMutex
	limiterMap     map[WorkStep]*rate.Limiter
	defaultLimiter *rate.Limiter
}

// GetDefaultLimiter get global limiter
func (l *limitManager) GetDefaultLimiter() *rate.Limiter {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.defaultLimiter
}

// SetDefaultLimiter set global limiter
func (l *limitManager) SetDefaultLimiter(r rate.Limit, b int) {
	limiter := rate.NewLimiter(r, b)

	l.mu.Lock()
	defer l.mu.Unlock()
	l.defaultLimiter = limiter
}

// GetLimiter get limiter for step
func (l *limitManager) GetLimiter(step WorkStep) *rate.Limiter {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if limiter, ok := l.limiterMap[step]; ok {
		return limiter
	} else {
		return l.defaultLimiter
	}
}

func (l *limitManager) SetLimiter(step WorkStep, r rate.Limit, b int) {
	limiter := rate.NewLimiter(r, b)

	l.mu.Lock()
	defer l.mu.Unlock()
	l.limiterMap[step] = limiter
}
