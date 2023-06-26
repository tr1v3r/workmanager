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

// NewPipeManager ...
func NewPipeManager(_ context.Context, steps ...WorkStep) *pipeManager { // nolint
	m := make(map[WorkStep]*pipe, len(steps))
	for _, step := range steps {
		ch := make(chan WorkTarget, defaultChanSize)
		m[step] = newPipe(ch, ch)
	}
	return &pipeManager{pipeMap: m}
}

func newPipe(recv <-chan WorkTarget, send chan<- WorkTarget) *pipe {
	return &pipe{recv: recv, send: send}
}

type pipe struct {
	recv <-chan WorkTarget
	send chan<- WorkTarget
}

func (p *pipe) Set(recv <-chan WorkTarget, send chan<- WorkTarget) {
	p.recv = recv
	p.send = send
}
func (p *pipe) SetRecv(recv <-chan WorkTarget) { p.recv = recv }
func (p *pipe) SetSend(send chan<- WorkTarget) { p.send = send }

type pipeManager struct {
	mu      sync.RWMutex
	pipeMap map[WorkStep]*pipe
}

func (pm *pipeManager) GetRecvChans(steps ...WorkStep) (chs []<-chan WorkTarget) {
	for _, step := range steps {
		chs = append(chs, pm.GetRecvChan(step))
	}
	return chs
}
func (pm *pipeManager) GetSendChans(steps ...WorkStep) (chs []chan<- WorkTarget) {
	for _, step := range steps {
		chs = append(chs, pm.GetSendChan(step))
	}
	return chs
}
func (pm *pipeManager) GetRecvChan(step WorkStep) <-chan WorkTarget { return pm.getPipe(step).recv }
func (pm *pipeManager) GetSendChan(step WorkStep) chan<- WorkTarget { return pm.getPipe(step).send }

func (pm *pipeManager) HasPipe(step WorkStep) bool { return pm.getPipe(step) != nil }

func (pm *pipeManager) SetPipe(step WorkStep, opts ...PipeOption) {
	ch := make(chan WorkTarget, defaultChanSize)
	for _, opt := range opts {
		ch = opt(ch)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.pipeMap[step] = newPipe(ch, ch)
}

func (pm *pipeManager) SetPipeChan(step WorkStep, recv <-chan WorkTarget, send chan<- WorkTarget) {
	if pipe := pm.pipeMap[step]; pipe != nil {
		pipe.Set(recv, send)
	}
}
func (pm *pipeManager) SetRecvChan(step WorkStep, recv <-chan WorkTarget) {
	if pipe := pm.pipeMap[step]; pipe != nil {
		pipe.SetRecv(recv)
	}
}
func (pm *pipeManager) SetSendChan(step WorkStep, send chan<- WorkTarget) {
	if pipe := pm.pipeMap[step]; pipe != nil {
		pipe.SetSend(send)
	}
}

func (pm *pipeManager) MITMSendChan(step WorkStep, newSendCh chan<- WorkTarget) chan<- WorkTarget {
	send := pm.GetSendChan(step)
	pm.SetSendChan(step, newSendCh)
	return send
}

func (pm *pipeManager) RemovePipe(steps ...WorkStep) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, step := range steps {
		delete(pm.pipeMap, step)
	}
}

func (pm *pipeManager) getPipe(step WorkStep) *pipe {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.pipeMap[step]
}
