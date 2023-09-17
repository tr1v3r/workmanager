package workmanager

import (
	"context"
	"runtime"
	"sync"

	"golang.org/x/time/rate"
)

const defaultBurst = 100
const defaultStepLimit rate.Limit = 100

func NewLimitController(_ context.Context, steps ...WorkStep) (mgr *limitController) {
	defer func() {
		for _, step := range steps {
			mgr.limiterMap[step] = rate.NewLimiter(defaultStepLimit, defaultBurst)
		}
	}()
	return &limitController{
		limiterMap:     make(map[WorkStep]*rate.Limiter),
		defaultLimiter: rate.NewLimiter(rate.Limit(runtime.NumCPU())*defaultStepLimit, defaultBurst),
	}
}

type limitController struct {
	mu             sync.RWMutex
	limiterMap     map[WorkStep]*rate.Limiter
	defaultLimiter *rate.Limiter
}

// getDefaultLimiter get global limiter
func (l *limitController) getDefaultLimiter() *rate.Limiter {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.defaultLimiter
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

func (l *limitController) SetLimiter(r rate.Limit, b int, steps ...WorkStep) {
	if len(steps) == 0 {
		return
	}
	if r < 0 {
		r = defaultStepLimit
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

func (l *limitController) limitSteps() (steps []WorkStep) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for step := range l.limiterMap {
		steps = append(steps, step)
	}
	return
}
