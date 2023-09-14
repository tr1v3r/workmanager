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
	Work(arg WorkTarget) ([]WorkTarget, error)
	AfterWork()

	GetResult() WorkTarget
	Finished() <-chan struct{}

	Terminate() error
}

// Cacher work target cache
type Cacher interface {
	// Allow to continue next steps when return true, abort step runner when return false
	Allow(tgt WorkTarget) bool
}

// WorkerConfig worker configure
type WorkerConfig interface {
	Args() map[string]interface{}
	Active() bool
}

// WorkTarget target/result
type WorkTarget interface {
	Token() string
	SetToken(token string)
	Key() string

	Trans(step WorkStep) ([]WorkTarget, error)

	TTL() int
}

// NewWorkTarget target/result
// TODO finish
type NewWorkTarget interface {
	// Token return target belong to which task
	Token() string
	// TTL return target time to live
	TTL() int
	// Clone return target clone
	Clone() NewWorkTarget
}

// WorkTask work task
type WorkTask interface {
	Cancel() error
	Finish() error

	IsCanceled() bool
	IsFinished() bool

	Token() string
	Context() context.Context
}
