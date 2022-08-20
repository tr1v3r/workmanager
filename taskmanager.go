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
	if task == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.tokenMap[task.Token()] = task
}

func (t *taskManager) GetTask(taskToken string) (task WorkTask) {
	defer func() {
		if task == nil {
			task = new(dummyTask)
		}
	}()

	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.tokenMap[taskToken]
}

func (t *taskManager) CancelTask(taskToken string) error { return t.GetTask(taskToken).Cancel() }
func (t *taskManager) TaskStart(taskToken string) error  { return t.GetTask(taskToken).Start() }
func (t *taskManager) TaskDone(taskToken string) error   { return t.GetTask(taskToken).Done() }
func (t *taskManager) IsCanceled(taskToken string) bool  { return t.GetTask(taskToken).IsCanceled() }

var _ WorkTask = new(dummyTask)

type dummyTask struct{}

func (*dummyTask) Start() error             { return nil }
func (*dummyTask) StartN(n int64) error     { return nil }
func (*dummyTask) Done() error              { return nil }
func (*dummyTask) Cancel() error            { return nil }
func (*dummyTask) IsCanceled() bool         { return false }
func (*dummyTask) IsFinished() bool         { return false }
func (*dummyTask) Token() string            { return "" }
func (*dummyTask) Context() context.Context { return nil }
