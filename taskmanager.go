package workmanager

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// NewTaskManager ...
func NewTaskManager(ctx context.Context) *taskManager { // nolint
	return &taskManager{
		ctx:      ctx,
		mu:       new(sync.RWMutex),
		tokenMap: make(map[string]*Task),
	}
}

type taskManager struct {
	ctx context.Context

	mu       *sync.RWMutex
	tokenMap map[string]*Task
}

func (t taskManager) WithContext(ctx context.Context) *taskManager {
	t.ctx = ctx
	return &t
}

func (t *taskManager) NewTask(step WorkStep) (task *Task, err error) {
	taskToken := uuid.New().String()
	c, cancel := context.WithCancel(t.ctx)

	task = &Task{
		Ctx:        c,
		TaskToken:  taskToken,
		StartTime:  time.Now(),
		CancelFunc: cancel,
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.tokenMap[taskToken] = task
	return
}

func (t *taskManager) Get(taskToken string) *Task {
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

func (t *taskManager) CancelTask(taskToken string) error {
	t.Get(taskToken).Cancel()
	return nil
}

func (t *taskManager) IsCanceled(taskToken string) bool {
	return t.Get(taskToken).IsCanceled()
}
