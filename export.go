package workmanager

import (
	"context"
)

type (
	// Work ...
	Work func(target WorkTarget, configs map[WorkerName]WorkerConfig) (results []WorkTarget, err error)

	// WorkerBuilder ...
	WorkerBuilder func(ctx context.Context, args map[string]interface{}) Worker

	// StepRunner ...
	StepRunner func(work Work, workTarget WorkTarget, nexts ...func(WorkTarget))

	// StepCallback ...
	StepCallback func(...WorkTarget) []WorkTarget
)

// Register register worker and step runner/processor
func Register(
	from WorkStep,
	runner StepRunner,
	workers map[WorkerName]WorkerBuilder,
	to ...WorkStep,
) {
	defaultWorkerMgr.Register(from, runner, workers, to...)
}

// RegisterWorker register worker
func RegisterWorker(name WorkerName, builder WorkerBuilder) {
	defaultWorkerMgr.RegisterWorker(name, builder)
}

// RegisterStep register step runner and processor
func RegisterStep(from WorkStep, runner StepRunner, to ...WorkStep) {
	defaultWorkerMgr.RegisterStep(from, runner, to...)
}

// RegisterBeforeCallbacks ...
func RegisterBeforeCallbacks(step WorkStep, callbacks ...StepCallback) {
	defaultWorkerMgr.RegisterBeforeCallbacks(step, callbacks...)
}

// RegisterAfterCallbacks ...
func RegisterAfterCallbacks(step WorkStep, callbacks ...StepCallback) {
	defaultWorkerMgr.RegisterAfterCallbacks(step, callbacks...)
}

// Serve daemon serve goroutine
func Serve(steps ...WorkStep) { defaultWorkerMgr.Serve(steps...) }

// Recv ...
func Recv(step WorkStep, target WorkTarget) error { return defaultWorkerMgr.Recv(step, target) }

// RecvFrom recv from chan
func RecvFrom(step WorkStep, recv <-chan WorkTarget) error {
	return defaultWorkerMgr.RecvFrom(step, recv)
}

// SetCacher set default work manager cacher
func SetCacher(c Cacher) { defaultWorkerMgr.SetCacher(c) }

// ListSteps list all steps
func ListSteps() []WorkStep { return defaultWorkerMgr.ListSteps() }

// PoolStastus return pool status
func PoolStatus(step WorkStep) (num, size int) { return defaultWorkerMgr.PoolStatus(step) }

// Task api

// AddTask ...
func AddTask(task WorkTask) { defaultWorkerMgr.AddTask(task) }

// GetTask ...
func GetTask(token string) WorkTask { return defaultWorkerMgr.GetTask(token) }

// CancelTask ...
func CancelTask(token string) error { return defaultWorkerMgr.CancelTask(token) }
