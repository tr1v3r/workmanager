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
	sz := runtime.NumCPU() * flex
	mgr := &poolManager{
		size: sz,
		m:    make(map[WorkStep]pools.Pool),
	}
	for _, step := range steps {
		mgr.m[step] = pools.NewPool(sz)
	}
	return mgr
}

type poolManager struct {
	size int
	mu   sync.RWMutex
	m    map[WorkStep]pools.Pool
}

func (p *poolManager) ResizePool(size int, steps ...WorkStep) error {
	for _, step := range steps {
		pool := pools.NewPool(size)
		p.mu.Lock()
		p.m[step] = pool
		p.mu.Unlock()
	}
	return nil
}

func (p *poolManager) Get(step WorkStep) pools.Pool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.m[step]
}

func (p *poolManager) Add(step WorkStep, size int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.m[step]; ok {
		return
	}
	if size >= 0 {
		size = runtime.NumCPU() * flex
	}
	p.m[step] = pools.NewPool(size)
}

func (p *poolManager) Remove(steps ...WorkStep) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, step := range steps {
		delete(p.m, step)
	}
}

// Size ...
func (p *poolManager) Size() int { return p.size }
