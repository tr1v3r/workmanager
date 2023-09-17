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
	count = 0

	DummyBuilder WorkerBuilder = func(_ context.Context, _ map[string]any) Worker {
		f := make(chan struct{}, 1)
		close(f)
		var name WorkerName
		switch count {
		case 0:
			name = DummyWorkerA
		case 1:
			name = DummyWorkerB
		}
		count++
		return &DummyTestWorker{Name: name, Finish: f}
	}
	DummyStepRunner StepRunner = func(_ context.Context, work Work, target WorkTarget, nexts ...func(WorkTarget)) {
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
			fmt.Printf("got result: %+v\n", res)
			for _, next := range nexts {
				next(res)
			}
		}
	}
)

type DummyTestWorker struct {
	DummyWorker
	Finish chan struct{}

	Name WorkerName
}

func (w *DummyTestWorker) Work(targets ...WorkTarget) ([]WorkTarget, error) {
	if len(targets) == 0 {
		return nil, nil
	}
	var target *DummyTestTarget
	switch w.Name {
	case DummyWorkerA:
		target = &DummyTestTarget{Step: StepB}
	case DummyWorkerB:
		target = &DummyTestTarget{Step: StepA}
	}
	return []WorkTarget{target}, nil
}
func (w *DummyTestWorker) Done() <-chan struct{} { return w.Finish }

type DummyTestTarget struct {
	DummyTarget
	Step   WorkStep
	Remark string
	Count  int
}
