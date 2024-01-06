package workmanager

import (
	"context"
	"sync"
)

// NewTaskController ...
func NewTaskController(ctx context.Context) *taskController { // nolint
	return &taskController{
		ctx:      ctx,
		mu:       new(sync.RWMutex),
		tokenMap: make(map[string]WorkTask),
	}
}

type taskController struct {
	ctx context.Context

	mu       *sync.RWMutex
	tokenMap map[string]WorkTask
}

func (t taskController) WithContext(ctx context.Context) *taskController {
	t.ctx = ctx
	return &t
}

func (t *taskController) AddTask(task WorkTask) {
	if task == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.tokenMap[task.Token()] = task
}

func (t *taskController) GetTask(taskToken string) (task WorkTask) {
	defer func() {
		if task == nil {
			task = new(dummyTask)
		}
	}()

	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.tokenMap[taskToken]
}

func (t *taskController) CancelTask(taskToken string) error { return t.GetTask(taskToken).Cancel() }
func (t *taskController) IsCanceled(taskToken string) bool  { return t.GetTask(taskToken).IsCanceled() }
func (t *taskController) FinishTask(taskToken string) error { return t.GetTask(taskToken).Finish() }
func (t *taskController) IsFinished(taskToken string) bool  { return t.GetTask(taskToken).IsFinished() }

var _ WorkTask = new(dummyTask)

type dummyTask struct{}

func (*dummyTask) Cancel() error            { return nil }
func (*dummyTask) Finish() error            { return nil }
func (*dummyTask) IsCanceled() bool         { return false }
func (*dummyTask) IsFinished() bool         { return false }
func (*dummyTask) Token() string            { return "" }
func (*dummyTask) Context() context.Context { return context.Background() }
