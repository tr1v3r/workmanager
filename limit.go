package workmanager

import (
	"context"
	"runtime"
	"sync"

	"golang.org/x/time/rate"
)

const (
	defaultBurst            = 100
	defaultLimit rate.Limit = 100
)

// NewLimitController create new limit controller
func NewLimitController(_ context.Context, steps ...WorkStep) (mgr *limitController) {
	defer func() {
		for _, step := range steps {
			mgr.limiterMap[step] = rate.NewLimiter(defaultLimit, defaultBurst)
		}
	}()
	return &limitController{
		limiterMap:     make(map[WorkStep]*rate.Limiter),
		defaultLimiter: rate.NewLimiter(rate.Limit(runtime.NumCPU())*defaultLimit, defaultBurst),
	}
}

type limitController struct {
	mu             sync.RWMutex
	limiterMap     map[WorkStep]*rate.Limiter
	defaultLimiter *rate.Limiter
}

// SetDefaultLimiter set global limiter
func (l *limitController) SetDefaultLimiter(r rate.Limit, b int) {
	limiter := rate.NewLimiter(r, b)

	l.mu.Lock()
	defer l.mu.Unlock()
	l.defaultLimiter = limiter
}

// getLimiter get limiter for step
func (l *limitController) getLimiter(step WorkStep) *rate.Limiter {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if limiter, ok := l.limiterMap[step]; ok {
		return limiter
	} else {
		return l.defaultLimiter
	}
}

func (l *limitController) SetLimit(r rate.Limit, b int, steps ...WorkStep) {
	if len(steps) == 0 {
		return
	}

	if r < 0 {
		r = defaultLimit
	}
	if b < 0 {
		b = defaultBurst
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	for _, step := range steps {
		if limiter, ok := l.limiterMap[step]; ok {
			limiter.SetLimit(r)
			limiter.SetBurst(b)
		} else {
			l.limiterMap[step] = rate.NewLimiter(r, b)
		}
	}
}
func (l *limitController) DelLimiter(steps ...WorkStep) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, step := range steps {
		delete(l.limiterMap, step)
	}
}

func (l *limitController) LimiterStatus(step WorkStep) (limit rate.Limit, burst int, tokens float64) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	limiter := l.getLimiter(step)
	return limiter.Limit(), limiter.Burst(), limiter.Tokens()
}

func (l *limitController) limitSteps() (steps []WorkStep) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for step := range l.limiterMap {
		steps = append(steps, step)
	}
	return steps
}
