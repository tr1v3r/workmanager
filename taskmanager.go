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

func (t *taskManager) AddTask(task WorkTask) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tokenMap[task.Token()] = task
}

func (t *taskManager) GetTask(taskToken string) WorkTask {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.tokenMap[taskToken]
}

func (t *taskManager) CancelTask(taskToken string) error { return t.GetTask(taskToken).Cancel() }

func (t *taskManager) TaskStart(taskToken string) { t.GetTask(taskToken).Start() }

func (t *taskManager) TaskDone(taskToken string) { t.GetTask(taskToken).Done() }

func (t *taskManager) IsCanceled(taskToken string) bool { return t.GetTask(taskToken).IsCanceled() }
