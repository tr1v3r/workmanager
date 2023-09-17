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
	// WithContext set worker context
	WithContext(context.Context) Worker

	// Work worker do work
	Work(targets ...WorkTarget) (results []WorkTarget, err error)
}

// Cacher work target cache
type Cacher interface {
	// Allow to continue next steps when return true, abort step runner when return false
	Allow(tgt WorkTarget) bool
}

// WorkerConfig worker configure
type WorkerConfig interface {
	Args() map[string]any
	Active() bool
}

// WorkTarget target/result
type WorkTarget interface {
	// Token return target belong to which task
	Token() string

	// Key return target unique key
	Key() string
	// TTL return target time to live
	TTL() int
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
