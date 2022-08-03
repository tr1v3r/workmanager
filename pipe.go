package workmanager

import (
	"context"
	"sync"
)

const defaultChanSize = 256

// NewPoolManager ...
func NewPipeManager(_ context.Context, steps ...WorkStep) *pipeManager { // nolint
	m := make(map[WorkStep]chan WorkTarget, len(steps))
	for _, step := range steps {
		m[step] = make(chan WorkTarget, defaultChanSize)
	}
	return &pipeManager{chanMap: m}
}

type pipeManager struct {
	mu      sync.RWMutex
	chanMap map[WorkStep]chan WorkTarget
}

func (pm *pipeManager) GetReadChan(step WorkStep) <-chan WorkTarget  { return pm.getChan(step) }
func (pm *pipeManager) GetWriteChan(step WorkStep) chan<- WorkTarget { return pm.getChan(step) }

func (pm *pipeManager) getChan(step WorkStep) chan WorkTarget {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.chanMap[step]
}

func (pm *pipeManager) GetReadChans(steps ...WorkStep) (chs []<-chan WorkTarget) {
	for _, step := range steps {
		if ch := pm.GetReadChan(step); ch != nil {
			chs = append(chs, ch)
		}
	}
	return chs
}
func (pm *pipeManager) GetWriteChans(steps ...WorkStep) (chs []chan<- WorkTarget) {
	for _, step := range steps {
		if ch := pm.GetWriteChan(step); ch != nil {
			chs = append(chs, ch)
		}
	}
	return chs
}

func (pm *pipeManager) Has(step WorkStep) bool {
	return pm.getChan(step) != nil
}

func (pm *pipeManager) SetStep(step WorkStep, opts ...StepOption) {
	ch := make(chan WorkTarget, defaultChanSize)
	for _, opt := range opts {
		ch = opt(ch)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.chanMap[step] = ch
}

func (pm *pipeManager) Remove(steps ...WorkStep) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, step := range steps {
		delete(pm.chanMap, step)
	}
}
