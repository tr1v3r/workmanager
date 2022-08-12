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
		f := make(chan struct{}, 1)
		close(f)
		return &dummyWorker{finish: f}
	}
	dummyStepRunner wm.StepRunner = func(work wm.Work, target wm.WorkTarget, nexts ...func(wm.WorkTarget)) {
		results, err := work(target, map[wm.WorkerName]wm.WorkerConfig{
			DummyWorkerA: new(wm.DummyConfig),
		})
		if err != nil {
			return
		}
		for _, next := range nexts {
			for _, res := range results {
				next(res)
			}
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
	_ = target.(*dummyTarget)
	return []wm.WorkTarget{&dummyTarget{step: StepB}}, nil
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
