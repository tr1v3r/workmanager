package workmanager

import (
	"sync"
)

func newCallbackController() *callbackController {
	return &callbackController{
		callbacks: make(map[WorkStep]*callbacks),
	}
}

type callbackController struct {
	mu        sync.RWMutex
	callbacks map[WorkStep]*callbacks
}

func (m *callbackController) GetCallbacks(step WorkStep) *callbacks {
	callback := m.getCallback(step)
	if callback == nil {
		callback = new(callbacks)
		m.setCallback(step, callback)
	}
	return callback
}

func (m *callbackController) getCallback(step WorkStep) *callbacks {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callbacks[step]
}

func (m *callbackController) setCallback(step WorkStep, ctr *callbacks) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callbacks[step] = ctr
}

type callbacks struct {
	mu     sync.RWMutex
	before []StepCallback
	after  []StepCallback
}

func (c *callbacks) RegisterBefore(calls ...StepCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.before = append(c.before, calls...)
}
func (c *callbacks) RegisterAfter(calls ...StepCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.after = append(c.after, calls...)
}
func (c *callbacks) BeforeWork() []StepCallback {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.before
}
func (c *callbacks) AfterWork() []StepCallback {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.after
}
