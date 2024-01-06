package workmanager

import (
	"context"
	"fmt"

	"github.com/tr1v3r/pkg/log"
)

const (
	DummyWorkerA WorkerName = "worker_a"
	DummyWorkerB WorkerName = "worker_b"

	StepA WorkStep = "step_a"
	StepB WorkStep = "step_b"
)

var (
	DummyBuilder func(WorkerName) WorkerBuilder = func(name WorkerName) WorkerBuilder {
		return func(_ context.Context, _ map[string]any) Worker {
			f := make(chan struct{}, 1)
			close(f)
			return &DummyTestWorker{Name: name, Finish: f}
		}
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
			log.Error("work fail: %s", err)
			return
		}

		for _, res := range results {
			fmt.Printf("[%s] got result: %+v\n", target.(*DummyTestTarget).Step, res)
			for _, next := range nexts {
				next(res)
			}
		}
	}
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

type DummyTestWorker struct {
	DummyWorker
	Finish chan struct{}

	Name WorkerName
}

func (w *DummyTestWorker) Work(targets ...WorkTarget) (results []WorkTarget, err error) {
	if len(targets) == 0 {
		return nil, nil
	}
	for _, target := range targets {
		switch w.Name {
		case DummyWorkerA:
			results = append(results, &DummyTestTarget{DummyTarget: target.(*DummyTestTarget).DummyTarget, Step: StepB})
		case DummyWorkerB:
			results = append(results, &DummyTestTarget{DummyTarget: target.(*DummyTestTarget).DummyTarget, Step: StepA})
		}
	}
	return
}
func (w *DummyTestWorker) Done() <-chan struct{} { return w.Finish }

type DummyTestTarget struct {
	DummyTarget
	Step   WorkStep
	Remark string
	Count  int
}
