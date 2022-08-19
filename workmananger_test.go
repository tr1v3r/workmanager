package workmanager

import (
	"context"
	"fmt"
)

func ContainsStep(step WorkStep, steps ...WorkStep) bool {
	if len(steps) == 0 {
		return false
	}

	for _, s := range steps {
		if s == step {
			return true
		}
	}
	return false
}

const (
	DummyWorkerA WorkerName = "worker_a"
	DummyWorkerB WorkerName = "worker_b"

	StepA WorkStep = "step_a"
	StepB WorkStep = "step_b"
)

var (
	DummyBuilder WorkerBuilder = func(_ context.Context, name WorkerName, _ map[string]interface{}) Worker {
		f := make(chan struct{}, 1)
		close(f)
		return &DummyTestWorker{Name: name, Finish: f}
	}
	DummyStepRunner StepRunner = func(work Work, target WorkTarget, nexts ...func(WorkTarget)) {
		var workerName WorkerName
		switch target.(*DummyTestTarget).Step {
		case StepA:
			workerName = DummyWorkerA
		case StepB:
			workerName = DummyWorkerB
		}
		results, err := work(target, map[WorkerName]WorkerConfig{
			workerName: new(DummyConfig),
		})
		if err != nil {
			return
		}

		for _, res := range results {
			for _, next := range nexts {
				next(res)
			}
		}
	}
	DummyStepProcessor StepProcessor = func(results ...WorkTarget) ([]WorkTarget, error) {
		for _, result := range results {
			fmt.Printf("got result: %+v\n", result)
		}
		return results, nil
	}
)

type DummyTestWorker struct {
	DummyWorker
	Finish chan struct{}

	Name WorkerName
}

func (w *DummyTestWorker) Work(target WorkTarget) ([]WorkTarget, error) {
	_ = target.(*DummyTestTarget)
	switch w.Name {
	case DummyWorkerA:
		target = &DummyTestTarget{Step: StepB}
	case DummyWorkerB:
		target = &DummyTestTarget{Step: StepA}
	}
	return []WorkTarget{target}, nil
}
func (w *DummyTestWorker) Finished() <-chan struct{} { return w.Finish }

type DummyTestTarget struct {
	DummyTarget
	Step WorkStep
}
