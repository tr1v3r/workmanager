package workmanager

import (
	"context"
)

// WorkerName worker name
type WorkerName string

// WorkStep work step
type WorkStep string

// Worker a worker
type Worker interface {
	LoadConfig(WorkerConfig) Worker
	WithContext(context.Context) Worker
	GetContext() context.Context

	BeforeWork()
	Work(arg WorkTarget) (WorkTarget, error)
	AfterWork()

	GetResult() WorkTarget
	Finished() <-chan struct{}

	Terminate() error
}

// WorkerConfig worker configure
type WorkerConfig interface {
	Args() map[string]interface{}
	Active() bool
}

// WorkTarget target/result
type WorkTarget interface {
	Token() string
	Key() string
	Step() WorkStep

	Trans(step WorkStep) (WorkTarget, error)
	ToArray() []WorkTarget
	Combine(...WorkTarget) WorkTarget

	TTL() int
}

// WorkTask work task
type WorkTask interface {
	Start()
	StartN(n int64)
	Done()

	Cancel()

	IsCanceled() bool
	IsFinished() bool

	Token() string
	Context() context.Context
}
