package workmanager

import (
	"context"

	"github.com/riverchu/pkg/log"
)

// Work ...
type Work func(target WorkTarget, configs map[WorkerName]WorkerConfig) (results []WorkTarget, err error)

// WorkerBuilder ...
type WorkerBuilder func(ctx context.Context, workerName WorkerName, args map[string]interface{}) Worker

// StepRunner ...
type StepRunner func(work Work, workTarget WorkTarget, to ...chan<- WorkTarget)

// StepProcessor ...
type StepProcessor func(results ...WorkTarget) error

// Register register worker and step runner/processor
func Register(
	from WorkStep,
	runner StepRunner,
	processor StepProcessor,
	workers map[WorkerName]WorkerBuilder,
	to ...WorkStep,
) {
	for name, builder := range workers {
		workerMgr.RegisterWorker(name, builder)
	}
	workerMgr.RegisterStep(from, runner, processor, to...)
}

// RegisterWorker register worker
func RegisterWorker(
	name WorkerName,
	builder WorkerBuilder,
) {
	workerMgr.RegisterWorker(name, builder)
}

// RegisterStep register step runner and processor
func RegisterStep(
	from WorkStep,
	runner StepRunner, // step runner
	processor StepProcessor, // result processor
	to ...WorkStep,
) {
	workerMgr.RegisterStep(from, runner, processor, to...)
}

// Serve ...
func Serve(steps ...WorkStep) { workerMgr.Serve(steps...) }

// Recv ...
func Recv(step WorkStep, target WorkTarget) error {
	log.Info("recv task, target is: %+v", target)
	return workerMgr.Recv(step, target)
}

// Task api

// NewTask ...
func NewTask(step WorkStep) (*Task, error) { return workerMgr.taskMgr.NewTask(step) }

// GetTask ...
func GetTask(token string) *Task { return workerMgr.taskMgr.Get(token) }

// CancelTask ...
func CancelTask(token string) error { return workerMgr.taskMgr.CancelTask(token) }
