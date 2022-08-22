package workmanager

import (
	"sync"
)

func newCallbackManager() *callbackManager {
	return &callbackManager{
		callbacks: make(map[WorkStep]*callbackController),
	}
}

type callbackManager struct {
	mu        sync.RWMutex
	callbacks map[WorkStep]*callbackController
}

func (m *callbackManager) GetCallbacks(step WorkStep) *callbackController {
	ctr := m.getCallbackCtr(step)
	if ctr == nil {
		ctr = new(callbackController)
		m.setCallbackCtr(step, ctr)
	}
	return ctr
}

func (m *callbackManager) getCallbackCtr(step WorkStep) *callbackController {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callbacks[step]
}

func (m *callbackManager) setCallbackCtr(step WorkStep, ctr *callbackController) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callbacks[step] = ctr
}

type callbackController struct {
	mu     sync.RWMutex
	before []StepCallback
	after  []StepCallback
}

func (t *callbackController) RegisterBefore(calls ...StepCallback) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.before = append(t.before, calls...)
}
func (t *callbackController) RegisterAfter(calls ...StepCallback) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.after = append(t.after, calls...)
}
func (t *callbackController) BeforeWork() []StepCallback {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.before
}
func (t *callbackController) AfterWork() []StepCallback {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.after
}
