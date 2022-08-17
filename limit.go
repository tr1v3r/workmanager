package workmanager

import (
	"context"
	"runtime"
	"sync"

	"golang.org/x/time/rate"
)

const defaultBurst = 100
const defaultStepLimit rate.Limit = 100

func NewLimitManager(_ context.Context, steps ...WorkStep) (mgr *limitManager) {
	defer func() {
		for _, step := range steps {
			mgr.limiterMap[step] = rate.NewLimiter(defaultStepLimit, defaultBurst)
		}
	}()
	return &limitManager{
		limiterMap:     make(map[WorkStep]*rate.Limiter),
		defaultLimiter: rate.NewLimiter(rate.Limit(runtime.NumCPU())*defaultStepLimit, defaultBurst),
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

func (l *limitManager) SetLimiter(r rate.Limit, steps ...WorkStep) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, step := range steps {
		if limiter, ok := l.limiterMap[step]; ok {
			limiter.SetLimit(r)
		} else {
			l.limiterMap[step] = rate.NewLimiter(r, defaultBurst)
		}
	}
}
func (l *limitManager) DelLimiter(steps ...WorkStep) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, step := range steps {
		delete(l.limiterMap, step)
	}
}

func (l *limitManager) LimitSteps() (steps []WorkStep) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for step := range l.limiterMap {
		steps = append(steps, step)
	}
	return
}
