package workmanager

import (
	"sync"
)

func newCallbackController() *callbackController {
	return &callbackController{callbacks: make(map[WorkStep]*callbacks)}
}

type callbackController struct {
	mu        sync.RWMutex
	callbacks map[WorkStep]*callbacks
}

// GetCallbacks query specified step's callbacks, if not found, build a new one and return it
func (ctr *callbackController) GetCallbacks(step WorkStep) *callbacks {
	c := ctr.getCallbacks(step)
	if c == nil {
		c = new(callbacks)
		ctr.setCallback(step, c)
	}
	return c
}

func (ctr *callbackController) CallbackStatus(step WorkStep) (beforeHookCount, afterHookCount int) {
	if callbacks := ctr.getCallbacks(step); callbacks == nil {
		return callbacks.Status()
	}
	return
}

func (ctr *callbackController) getCallbacks(step WorkStep) *callbacks {
	ctr.mu.RLock()
	defer ctr.mu.RUnlock()
	return ctr.callbacks[step]
}

func (ctr *callbackController) setCallback(step WorkStep, callbacks *callbacks) {
	ctr.mu.Lock()
	defer ctr.mu.Unlock()
	ctr.callbacks[step] = callbacks
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
func (c *callbacks) Status() (beforeHookCount, afterHookCount int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.before), len(c.after)
}
