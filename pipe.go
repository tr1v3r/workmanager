package workmanager

import (
	"context"
	"reflect"
	"sync"
)

const defaultChanSize = 256

// PipeOption pipe option
type PipeOption func(chan WorkTarget) chan WorkTarget

// PipeChSize set pipe channle size
var PipeChSize = func(size int) PipeOption {
	return func(ch chan WorkTarget) chan WorkTarget {
		if size < 0 {
			return ch
		}
		return make(chan WorkTarget, size)
	}
}

// NewPipeController ...
func NewPipeController(_ context.Context, steps ...WorkStep) *pipeController { // nolint
	m := make(map[WorkStep]*pipe, len(steps))
	for _, step := range steps {
		ch := make(chan WorkTarget, defaultChanSize)
		m[step] = newPipe(ch, ch)
	}
	return &pipeController{pipeMap: m}
}

func newPipe(recv <-chan WorkTarget, send chan<- WorkTarget) *pipe {
	return &pipe{recv: recv, send: send}
}

type pipe struct {
	recv <-chan WorkTarget
	send chan<- WorkTarget
}

func (p *pipe) Set(recv <-chan WorkTarget, send chan<- WorkTarget) {
	p.SetRecv(recv)
	p.SetSend(send)
}
func (p *pipe) SetRecv(recv <-chan WorkTarget) { p.recv = recv }
func (p *pipe) SetSend(send chan<- WorkTarget) { p.send = send }

func (p *pipe) Status() (same bool, recvLen, recvCap, sendLen, sendCap int) {
	return reflect.ValueOf(p.recv).Pointer() == reflect.ValueOf(p.send).Pointer(),
		len(p.recv), cap(p.recv),
		len(p.send), cap(p.send)
}

type pipeController struct {
	mu      sync.RWMutex
	pipeMap map[WorkStep]*pipe
}

func (ctr *pipeController) HasPipe(step WorkStep) bool { return ctr.getPipe(step) != nil }

func (ctr *pipeController) InitializePipe(step WorkStep, opts ...PipeOption) *pipe {
	ch := make(chan WorkTarget, defaultChanSize)
	for _, opt := range opts {
		ch = opt(ch)
	}

	pipe := newPipe(ch, ch)

	ctr.mu.Lock()
	defer ctr.mu.Unlock()
	ctr.pipeMap[step] = pipe

	return pipe
}

func (ctr *pipeController) GetRecvChans(steps ...WorkStep) (chs []<-chan WorkTarget) {
	for _, step := range steps {
		chs = append(chs, ctr.GetRecvChan(step))
	}
	return chs
}
func (ctr *pipeController) GetSendChans(steps ...WorkStep) (chs []chan<- WorkTarget) {
	for _, step := range steps {
		chs = append(chs, ctr.GetSendChan(step))
	}
	return chs
}
func (ctr *pipeController) GetRecvChan(step WorkStep) <-chan WorkTarget {
	if pipe := ctr.getPipe(step); pipe != nil {
		return pipe.recv
	}
	return nil
}
func (ctr *pipeController) GetSendChan(step WorkStep) chan<- WorkTarget {
	if pipe := ctr.getPipe(step); pipe != nil {
		return pipe.send
	}
	return nil
}

func (ctr *pipeController) SetPipeChan(step WorkStep, recv <-chan WorkTarget, send chan<- WorkTarget) {
	if pipe := ctr.getPipe(step); pipe != nil { // if found pipe for step
		pipe.Set(recv, send)
	} else { // if not found, init one and set recv and send
		ctr.InitializePipe(step).Set(recv, send)
	}
}
func (ctr *pipeController) SetRecvChan(step WorkStep, recv <-chan WorkTarget) {
	if pipe := ctr.getPipe(step); pipe != nil {
		pipe.SetRecv(recv)
	} else {
		ctr.InitializePipe(step).SetRecv(recv)
	}
}
func (ctr *pipeController) SetSendChan(step WorkStep, send chan<- WorkTarget) {
	if pipe := ctr.getPipe(step); pipe != nil {
		pipe.SetSend(send)
	} else {
		ctr.InitializePipe(step).SetSend(send)
	}
}

func (ctr *pipeController) MITMSendChan(step WorkStep, send chan<- WorkTarget) chan<- WorkTarget {
	originSend := ctr.GetSendChan(step)
	ctr.SetSendChan(step, send)
	return originSend
}

func (ctr *pipeController) RemovePipe(steps ...WorkStep) {
	ctr.mu.Lock()
	defer ctr.mu.Unlock()
	for _, step := range steps {
		delete(ctr.pipeMap, step)
	}
}

func (ctr *pipeController) PipeStatus(step WorkStep) (same bool, recvLen, recvCap, sendLen, sendCap int) {
	if pipe := ctr.getPipe(step); pipe != nil {
		return pipe.Status()
	}
	return
}

func (ctr *pipeController) getPipe(step WorkStep) *pipe {
	ctr.mu.RLock()
	defer ctr.mu.RUnlock()
	return ctr.pipeMap[step]
}
