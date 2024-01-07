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

func newPipeWithOpts(opts ...PipeOption) *pipe {
	ch := make(chan WorkTarget, defaultChanSize)
	for _, opt := range opts {
		ch = opt(ch)
	}
	return newPipe(ch, ch)
}

////////////////////////////////////////////////////////////
//
// code for pipe struct
//
////////////////////////////////////////////////////////////

// pipe used for trans data between steps
type pipe struct {
	// recv recv chan for step
	recv <-chan WorkTarget
	// send send chan for step, the data sent to this chan will be consumed by recv chan
	send chan<- WorkTarget
}

func (p *pipe) set(recv <-chan WorkTarget, send chan<- WorkTarget) {
	p.setRecv(recv)
	p.setSend(send)
}
func (p *pipe) setRecv(recv <-chan WorkTarget) { p.recv = recv }
func (p *pipe) setSend(send chan<- WorkTarget) { p.send = send }

func (p *pipe) check() error {
	if p == nil {
		return ErrPipeNotFound
	}
	if p.recv == nil {
		return ErrPipeRecvChanNil
	}
	if p.send == nil {
		return ErrPipeSendChanNil
	}
	return nil
}

func (p *pipe) status() (mitm bool, recvLen, recvCap, sendLen, sendCap int) {
	return reflect.ValueOf(p.recv).Pointer() != reflect.ValueOf(p.send).Pointer(),
		len(p.recv), cap(p.recv),
		len(p.send), cap(p.send)
}

////////////////////////////////////////////////////////////
//
// code for pipe controller
//
////////////////////////////////////////////////////////////

type pipeController struct {
	mu      sync.RWMutex
	pipeMap map[WorkStep]*pipe
}

// InitializePipe initialize pipe for step
// build a new chan for recv and send to step
func (ctr *pipeController) InitializePipe(step WorkStep, opts ...PipeOption) *pipe {
	pipe := newPipeWithOpts(opts...)
	ctr.setPipe(step, pipe)
	return pipe
}

// SetPipe set pipe for step
func (ctr *pipeController) SetPipe(step WorkStep, opts ...PipeOption) {
	ctr.setPipe(step, newPipeWithOpts(opts...))
}

// MITMSendChan set mitm send chan for step
func (ctr *pipeController) MITMSendChan(step WorkStep, send chan<- WorkTarget) chan<- WorkTarget {
	originSend := ctr.getSendChan(step)
	ctr.setSendChan(step, send)
	return originSend
}

// RemovePipe remove steps' pipe
func (ctr *pipeController) RemovePipe(steps ...WorkStep) {
	ctr.mu.Lock()
	defer ctr.mu.Unlock()
	for _, step := range steps {
		delete(ctr.pipeMap, step)
	}
}

// PipeStatus return pipe status
func (ctr *pipeController) PipeStatus(step WorkStep) (mitm bool, recvLen, recvCap, sendLen, sendCap int) {
	if pipe := ctr.getPipe(step); pipe != nil {
		return pipe.status()
	}
	return
}

func (ctr *pipeController) CheckPipe(step WorkStep) error {
	if pipe := ctr.getPipe(step); pipe != nil {
		return pipe.check()
	}
	return ErrPipeNotFound
}

func (ctr *pipeController) hasPipe(step WorkStep) bool { return ctr.getPipe(step) != nil }

func (ctr *pipeController) getRecvChans(steps ...WorkStep) (chs []<-chan WorkTarget) {
	for _, step := range steps {
		chs = append(chs, ctr.getRecvChan(step))
	}
	return chs
}
func (ctr *pipeController) getSendChans(steps ...WorkStep) (chs []chan<- WorkTarget) {
	for _, step := range steps {
		chs = append(chs, ctr.getSendChan(step))
	}
	return chs
}
func (ctr *pipeController) getRecvChan(step WorkStep) <-chan WorkTarget {
	if pipe := ctr.getPipe(step); pipe != nil {
		return pipe.recv
	}
	return nil
}
func (ctr *pipeController) getSendChan(step WorkStep) chan<- WorkTarget {
	if pipe := ctr.getPipe(step); pipe != nil {
		return pipe.send
	}
	return nil
}

func (ctr *pipeController) setPipeChan(step WorkStep, recv <-chan WorkTarget, send chan<- WorkTarget) {
	if pipe := ctr.getPipe(step); pipe != nil { // if found pipe for step
		pipe.set(recv, send)
	} else { // if not found, init one and set recv and send
		ctr.InitializePipe(step).set(recv, send)
	}
}
func (ctr *pipeController) setRecvChan(step WorkStep, recv <-chan WorkTarget) {
	if pipe := ctr.getPipe(step); pipe != nil {
		pipe.setRecv(recv)
	} else {
		ctr.InitializePipe(step).setRecv(recv)
	}
}
func (ctr *pipeController) setSendChan(step WorkStep, send chan<- WorkTarget) {
	if pipe := ctr.getPipe(step); pipe != nil {
		pipe.setSend(send)
	} else {
		ctr.InitializePipe(step).setSend(send)
	}
}

func (ctr *pipeController) setPipe(step WorkStep, pipe *pipe) {
	ctr.mu.Lock()
	defer ctr.mu.Unlock()
	ctr.pipeMap[step] = pipe
}
func (ctr *pipeController) getPipe(step WorkStep) *pipe {
	ctr.mu.RLock()
	defer ctr.mu.RUnlock()
	return ctr.pipeMap[step]
}
