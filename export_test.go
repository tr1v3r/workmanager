package workmanager_test

import (
	"context"
	"fmt"

	wm "github.com/riverchu/workmanager"
)

const (
	DummyWorkerA wm.WorkerName = "worker_a"
	DummyWorkerB wm.WorkerName = "worker_b"

	StepA wm.WorkStep = "step_a"
	StepB wm.WorkStep = "step_b"
)

var (
	dummyBuilder wm.WorkerBuilder = func(_ context.Context, name wm.WorkerName, _ map[string]interface{}) wm.Worker {
		f := make(chan struct{}, 1)
		close(f)
		return &dummyWorker{name: name, finish: f}
	}
	dummyStepRunner wm.StepRunner = func(work wm.Work, target wm.WorkTarget, nexts ...func(wm.WorkTarget)) {
		var workerName wm.WorkerName
		switch target.(*dummyTarget).step {
		case StepA:
			workerName = DummyWorkerA
		case StepB:
			workerName = DummyWorkerB
		}
		results, err := work(target, map[wm.WorkerName]wm.WorkerConfig{
			workerName: new(wm.DummyConfig),
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
	dummyStepProcessor wm.StepProcessor = func(results ...wm.WorkTarget) ([]wm.WorkTarget, error) {
		for _, result := range results {
			fmt.Printf("got result: %+v\n", result)
		}
		return results, nil
	}
)

type dummyWorker struct {
	wm.DummyWorker
	finish chan struct{}

	name wm.WorkerName
}

func (w *dummyWorker) Work(target wm.WorkTarget) ([]wm.WorkTarget, error) {
	_ = target.(*dummyTarget)
	switch w.name {
	case DummyWorkerA:
		target = &dummyTarget{step: StepB}
	case DummyWorkerB:
		target = &dummyTarget{step: StepA}
	}
	return []wm.WorkTarget{target}, nil
}
func (w *dummyWorker) Finished() <-chan struct{} { return w.finish }

type dummyTarget struct {
	wm.DummyTarget
	step wm.WorkStep
}
