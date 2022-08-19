package workmanager

import (
	"context"
)

type (
	// Work ...
	Work func(target WorkTarget, configs map[WorkerName]WorkerConfig) (results []WorkTarget, err error)

	// WorkerBuilder ...
	WorkerBuilder func(ctx context.Context, workerName WorkerName, args map[string]interface{}) Worker

	// StepRunner ...
	StepRunner func(work Work, workTarget WorkTarget, nexts ...func(WorkTarget))

	// StepProcessor ...
	StepProcessor func(results ...WorkTarget) ([]WorkTarget, error)
)

// Register register worker and step runner/processor
func Register(
	from WorkStep,
	runner StepRunner,
	processor StepProcessor,
	workers map[WorkerName]WorkerBuilder,
	to ...WorkStep,
) {
	defaultWorkerMgr.Register(from, runner, processor, workers, to...)
}

// RegisterWorker register worker
func RegisterWorker(
	name WorkerName,
	builder WorkerBuilder,
) {
	defaultWorkerMgr.RegisterWorker(name, builder)
}

// RegisterStep register step runner and processor
func RegisterStep(
	from WorkStep,
	runner StepRunner, // step runner
	processor StepProcessor, // result processor
	to ...WorkStep,
) {
	defaultWorkerMgr.RegisterStep(from, runner, processor, to...)
}

// Serve daemon serve goroutine
func Serve(steps ...WorkStep) { defaultWorkerMgr.Serve(steps...) }

// Recv ...
func Recv(step WorkStep, target WorkTarget) error { return defaultWorkerMgr.Recv(step, target) }

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
