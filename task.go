package workmanager

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

var _ WorkTask = new(Task)

// NewTask create new task
func NewTask(ctx context.Context) WorkTask {
	c, cancel := context.WithCancel(ctx)
	return &Task{
		Ctx:        c,
		TaskToken:  uuid.New().String(),
		StartTime:  time.Now(),
		CancelFunc: cancel,
	}
}

// Task ...
type Task struct {
	Ctx context.Context

	TaskToken string
	StartTime time.Time
	EndTime   time.Time

	CancelFunc context.CancelFunc

	FinishFunc func()
	Finished   bool

	startCount int64
	doneCount  int64
}

// Token ...
func (t *Task) Token() string { return t.TaskToken }

// Context ...
func (t *Task) Context() context.Context { return t.Ctx }

// Start ...
func (t *Task) Start() {
	if t == nil {
		return
	}
	atomic.AddInt64(&t.startCount, 1)
}

// StartN ...
func (t *Task) StartN(n int64) {
	if t == nil {
		return
	}
	atomic.AddInt64(&t.startCount, n)
} // nolint

// Done ...
func (t *Task) Done() {
	if t == nil {
		return
	}
	defer atomic.AddInt64(&t.doneCount, 1)
	if atomic.LoadInt64(&t.startCount) == atomic.LoadInt64(&t.doneCount)+1 {
		t.Finished = true
		t.EndTime = time.Now()
		if t.FinishFunc != nil {
			t.FinishFunc()
		}
	}
}

// Cancel ...
func (t *Task) Cancel() {
	if t == nil {
		return
	}
	if t.CancelFunc != nil {
		t.CancelFunc()
	}
}

// IsCanceled ...
func (t *Task) IsCanceled() bool {
	if t == nil {
		return true
	}
	return errors.Is(t.Ctx.Err(), context.Canceled)
}

// IsFinished ...
func (t *Task) IsFinished() bool {
	if t == nil {
		return true
	}
	return t.Finished
}
