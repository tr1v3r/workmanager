package workmanager

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

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

func (t *Task) Start() {
	if t == nil {
		return
	}
	atomic.AddInt64(&t.startCount, 1)
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

func (t *Task) AddScanStatus(status ...WorkStep) {
	if t == nil {
		return
	}
}
