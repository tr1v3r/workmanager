package workmanager

import (
	"context"
	"sync"
)

const defaultChanSize = 256

// PipeOption ...
type PipeOption func(chan WorkTarget) chan WorkTarget

// PipeChSize ...
var PipeChSize = func(size int) PipeOption {
	return func(ch chan WorkTarget) chan WorkTarget {
		if size < 0 {
			return ch
		}
		return make(chan WorkTarget, size)
	}
}

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

func (pm *pipeManager) GetRecvChans(steps ...WorkStep) (chs []<-chan WorkTarget) {
	for _, step := range steps {
		if ch := pm.GetRecvChan(step); ch != nil {
			chs = append(chs, ch)
		}
	}
	return chs
}
func (pm *pipeManager) GetSendChans(steps ...WorkStep) (chs []chan<- WorkTarget) {
	for _, step := range steps {
		if ch := pm.GetSendChan(step); ch != nil {
			chs = append(chs, ch)
		}
	}
	return chs
}
func (pm *pipeManager) GetRecvChan(step WorkStep) <-chan WorkTarget { return pm.getChan(step) }
func (pm *pipeManager) GetSendChan(step WorkStep) chan<- WorkTarget { return pm.getChan(step) }
func (pm *pipeManager) getChan(step WorkStep) chan WorkTarget {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.chanMap[step]
}

func (pm *pipeManager) HasPipe(step WorkStep) bool { return pm.getChan(step) != nil }
func (pm *pipeManager) SetPipe(step WorkStep, opts ...PipeOption) {
	ch := make(chan WorkTarget, defaultChanSize)
	for _, opt := range opts {
		ch = opt(ch)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.chanMap[step] = ch
}
func (pm *pipeManager) DelPipe(steps ...WorkStep) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, step := range steps {
		delete(pm.chanMap, step)
	}
}
