package workmanager

import (
	"context"
	"errors"
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

	FinishFunc func() error
	Finished   bool
}

// Token return task token
func (t *Task) Token() string { return t.TaskToken }

// Context return context
func (t *Task) Context() context.Context { return t.Ctx }

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

// Finish mark task finished
func (t *Task) Finish() error {
	if t == nil {
		return ErrNilTask
	}
	if t.FinishFunc != nil {
		return t.FinishFunc()
	}
	t.Finished = true
	return nil
}

// IsCanceled check if task canceled
func (t *Task) IsCanceled() bool { return t == nil || errors.Is(t.Ctx.Err(), context.Canceled) }

// IsFinished check if task finished
func (t *Task) IsFinished() bool { return t == nil || t.Finished }
