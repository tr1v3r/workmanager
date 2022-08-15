package workmanager

import (
	"context"
	"sync"
)

// NewTaskManager ...
func NewTaskManager(ctx context.Context) *taskManager { // nolint
	return &taskManager{
		ctx:      ctx,
		mu:       new(sync.RWMutex),
		tokenMap: make(map[string]WorkTask),
	}
}

type taskManager struct {
	ctx context.Context

	mu       *sync.RWMutex
	tokenMap map[string]WorkTask
}

func (t taskManager) WithContext(ctx context.Context) *taskManager {
	t.ctx = ctx
	return &t
}

func (t *taskManager) Add(task WorkTask) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tokenMap[task.Token()] = task
}

func (t *taskManager) Get(taskToken string) WorkTask {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.tokenMap[taskToken]
}

func (t *taskManager) Start(taskToken string) {
	t.Get(taskToken).Start()
}

func (t *taskManager) Done(taskToken string) {
	t.Get(taskToken).Done()
}

func (t *taskManager) Cancel(taskToken string) error {
	t.Get(taskToken).Cancel()
	return nil
}

func (t *taskManager) IsCanceled(taskToken string) bool {
	return t.Get(taskToken).IsCanceled()
}
