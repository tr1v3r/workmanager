package workmanager

import (
	"context"
	"runtime"
	"sync"

	"github.com/tr1v3r/pkg/pools"
)

const flex = 1

var defaultPoolSize = runtime.NumCPU() * flex

// NewPoolManager ...
func NewPoolManager(_ context.Context, steps ...WorkStep) (mgr *poolManager) { // nolint
	defer func() {
		for _, step := range steps {
			mgr.poolMap[step] = pools.NewPool(defaultPoolSize)
		}
	}()
	return &poolManager{
		poolMap:     make(map[WorkStep]pools.Pool),
		defaultPool: pools.NewPool(defaultPoolSize * len(steps)),
	}
}

type poolManager struct {
	mu          sync.RWMutex
	poolMap     map[WorkStep]pools.Pool
	defaultPool pools.Pool
}

// poolSteps return all step has pool
func (p *poolManager) poolSteps() (steps []WorkStep) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for step := range p.poolMap {
		steps = append(steps, step)
	}
	return
}

// getDefaultPool get default pool
func (p *poolManager) getDefaultPool() pools.Pool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.defaultPool
}

// SetDefaultPool set default pool
func (p *poolManager) SetDefaultPool(size int) {
	pool := pools.NewPool(size)

	p.mu.Lock()
	defer p.mu.Unlock()
	p.defaultPool = pool
}

func (p *poolManager) getPool(step WorkStep) pools.Pool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if pool, ok := p.poolMap[step]; ok {
		return pool
	} else {
		return p.defaultPool
	}
}

func (p *poolManager) SetPool(size int, steps ...WorkStep) {
	if len(steps) == 0 {
		return
	}

	if size <= 0 {
		size = defaultPoolSize
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	for _, step := range steps {
		p.poolMap[step] = pools.NewPool(size)
	}
}

func (p *poolManager) RemovePool(steps ...WorkStep) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, step := range steps {
		delete(p.poolMap, step)
	}
}

func (p *poolManager) PoolStatus(step WorkStep) (num, size int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.poolMap[step].Num(), p.poolMap[step].Size()
}
