package workmanager_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	wm "github.com/riverchu/workmanager"
)

const (
	DummyWorkerA wm.WorkerName = "worker_a"
	DummyWorkerB wm.WorkerName = "worker_b"

	StepA wm.WorkStep = "step_a"
	StepB wm.WorkStep = "step_b"
)

var (
	dummyBuilder wm.WorkerBuilder = func(context.Context, wm.WorkerName, map[string]interface{}) wm.Worker {
		return &dummyWorker{finish: make(chan struct{}, 1)}
	}
	dummyStepRunner wm.StepRunner = func(work wm.Work, target wm.WorkTarget, nexts ...func(wm.WorkTarget)) {
		_, err := work(target, map[wm.WorkerName]wm.WorkerConfig{
			DummyWorkerA: new(wm.DummyConfig),
			DummyWorkerB: new(wm.DummyConfig),
		})
		if err != nil {
			return
		}
		for _, next := range nexts {
			next(&dummyTarget{DummyTarget: wm.DummyTarget{TaskToken: target.Token()}, step: StepB})
		}
	}
	dummyStepProcessor wm.StepProcessor = func(results ...wm.WorkTarget) ([]wm.WorkTarget, error) {
		fmt.Printf("got result: %+v\n", results[0])
		return results, nil
	}
)

type dummyWorker struct {
	wm.DummyWorker
	finish chan struct{}
}

func (w *dummyWorker) Work(target wm.WorkTarget) ([]wm.WorkTarget, error) {
	return []wm.WorkTarget{target}, nil
}
func (w *dummyWorker) Finished() <-chan struct{} { return w.finish }

type dummyTarget struct {
	wm.DummyTarget
	step wm.WorkStep
}

func Test_Outline(t *testing.T) {
	wm.RegisterWorker(DummyWorkerA, dummyBuilder)
	wm.RegisterWorker(DummyWorkerB, dummyBuilder)

	wm.RegisterStep(StepA, dummyStepRunner, dummyStepProcessor, StepB)
	wm.RegisterStep(StepB, dummyStepRunner, dummyStepProcessor)

	wm.Serve()

	task := wm.NewTask(context.Background())
	wm.AddTask(task)

	err := wm.Recv(StepA, &dummyTarget{DummyTarget: wm.DummyTarget{TaskToken: task.Token()}, step: StepA})
	if err != nil {
		t.Errorf("send target fail: %s", err)
	}

	<-time.NewTimer(3 * time.Second).C

	t.Logf("task %+v", wm.GetTask(task.Token()))
}
