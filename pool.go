package workmanager

import (
	"context"
	"runtime"
	"sync"

	"github.com/riverchu/pkg/pools"
)

const flex = 1

// NewPoolManager ...
func NewPoolManager(_ context.Context, steps ...WorkStep) *poolManager { // nolint
	mgr := &poolManager{
		size: runtime.NumCPU() * flex,
		m:    make(map[WorkStep]pools.Pool),
	}
	for _, step := range steps {
		mgr.m[step] = pools.NewPool(mgr.size)
	}
	return mgr
}

type poolManager struct {
	size int
	mu   sync.RWMutex
	m    map[WorkStep]pools.Pool
}

// PoolSteps return all step has pool
func (p *poolManager) PoolSteps() (steps []WorkStep) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for step := range p.m {
		steps = append(steps, step)
	}
	return
}

func (p *poolManager) GetPool(step WorkStep) pools.Pool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.m[step]
}

func (p *poolManager) SetPool(size int, steps ...WorkStep) {
	if len(steps) == 0 {
		return
	}

	if size <= 0 {
		size = runtime.NumCPU() * flex
	}

	poolArr := make([]pools.Pool, len(steps))
	for i := range steps {
		poolArr[i] = pools.NewPool(size)
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	for i, step := range steps {
		p.m[step] = poolArr[i]
	}
}

func (p *poolManager) DelPool(steps ...WorkStep) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, step := range steps {
		delete(p.m, step)
	}
}

// Size ...
func (p *poolManager) Size() int { return p.size }
