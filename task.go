package workmanager

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var _ WorkTask = new(Task)

// NewTask create new task
func NewTask(ctx context.Context) WorkTask {
	c, cancel := context.WithCancel(ctx)
	return &Task{
		ctx:        c,
		taskToken:  uuid.New().String(),
		startTime:  time.Now(),
		cancelFunc: cancel,
	}
}

// Task work task
type Task struct {
	ctx        context.Context
	cancelFunc context.CancelFunc

	taskToken string
	startTime time.Time
	endTime   time.Time

	mu         sync.RWMutex
	finished   bool
	finishFunc func() error
}

// Token return task token
func (t *Task) Token() string { return t.taskToken }

// Context return context
func (t *Task) Context() context.Context { return t.ctx }

// Cancel task cancel
func (t *Task) Cancel() error {
	if t == nil {
		return ErrNilTask
	}

	t.finish()

	return t.callFinishFunc()
}

// Finish mark task finished
func (t *Task) Finish() error {
	if t == nil {
		return ErrNilTask
	}

	t.finish()

	return t.callFinishFunc()
}

func (t *Task) SetFinishFunc(f func() error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.finishFunc = f
}
func (t *Task) callFinishFunc() error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.finishFunc != nil {
		return t.finishFunc()
	}
	return nil
}

// IsCanceled check if task canceled
func (t *Task) IsCanceled() bool { return t == nil || errors.Is(t.ctx.Err(), context.Canceled) }

// IsFinished check if task finished
func (t *Task) IsFinished() bool { return t == nil || t.getFinished() }

func (t *Task) getFinished() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.finished
}
func (t *Task) finish() {
	if t.IsFinished() {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.finished = true
	t.endTime = time.Now()
}
