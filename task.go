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

// Task work task
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

// Token return task token
func (t *Task) Token() string { return t.TaskToken }

// Context return context
func (t *Task) Context() context.Context { return t.Ctx }

// Start task start
func (t *Task) Start() error { return t.StartN(1) }

// StartN task start n
func (t *Task) StartN(n int64) error {
	if t == nil {
		return ErrNilTask
	}
	atomic.AddInt64(&t.startCount, n)
	return nil
} // nolint

// Done task done count
func (t *Task) Done() error {
	if t == nil {
		return ErrNilTask
	}
	defer atomic.AddInt64(&t.doneCount, 1)
	if atomic.LoadInt64(&t.startCount) == atomic.LoadInt64(&t.doneCount)+1 {
		t.Finished = true
		t.EndTime = time.Now()
		if t.FinishFunc != nil {
			t.FinishFunc()
		}
	}
	return nil
}

// Cancel task cancel
func (t *Task) Cancel() error {
	if t == nil {
		return ErrNilTask
	}
	if t.CancelFunc != nil {
		t.CancelFunc()
	}
	return nil
}

// IsCanceled ...
func (t *Task) IsCanceled() bool { return t == nil || errors.Is(t.Ctx.Err(), context.Canceled) }

// IsFinished ...
func (t *Task) IsFinished() bool { return t == nil || t.Finished }
