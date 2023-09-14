package workmanager_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tr1v3r/pkg/log"
	wm "github.com/tr1v3r/workmanager"
)

func TestWork(t *testing.T) {
	mgr := wm.NewWorkerManager(context.Background())

	mgr.RegisterStep("step_a", func(ctx context.Context, work wm.Work, target wm.WorkTarget, nexts ...func(wm.WorkTarget)) {
		if err := ctx.Err(); err != nil {
			return
		}
		results, err := work(target, nil)
		if err != nil {
			t.Errorf("step_a work fail: %s", err)
			return
		}
		for _, result := range results {
			for _, next := range nexts {
				next(result)
			}
		}
	}, "step_b")
	mgr.RegisterStep("step_b", func(ctx context.Context, work wm.Work, target wm.WorkTarget, _ ...func(wm.WorkTarget)) {
		if err := ctx.Err(); err != nil {
			return
		}
		results, err := work(target, nil)
		if err != nil {
			t.Errorf("step_b work fail: %s", err)
			return
		}
		for _, result := range results {
			fmt.Printf("result: %+v\n", result)
		}
	})

	// mgr.SetPipe("step_a", wm.PipeChSize(8))

	mgr.Serve("step_a", "step_b")

	task := wm.NewTask(context.Background())
	task.(*wm.Task).TaskToken = "example_token_123"
	mgr.AddTask(task)

	err := mgr.Recv("step_a", &wm.DummyTestTarget{DummyTarget: wm.DummyTarget{TaskToken: task.Token()}, Step: "step_a"})
	if err != nil {
		fmt.Printf("send target fail: %s", err)
	}

	time.Sleep(10 * time.Second)
	// time.Sleep(100 * time.Millisecond)
}

var printer = new(Printer)

func PrinterBuilder(ctx context.Context, args map[string]interface{}) wm.Worker { return printer }

type Printer struct {
	counter int
	wm.DummyWorker
}

func (p *Printer) Work(target wm.WorkTarget) ([]wm.WorkTarget, error) {
	p.counter++

	log.Info("[%d] printer working", p.counter)
	if t, ok := target.(*wm.DummyTestTarget); ok {
		t := *t
		t.Remark += fmt.Sprintf("<%d>", p.counter)
		return []wm.WorkTarget{&t}, nil
	}
	return nil, nil
}
func (p *Printer) Finished() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestPrintWorker(t *testing.T) {
	mgr := wm.NewWorkerManager(context.Background())

	mgr.Register("print", func(ctx context.Context, work wm.Work, target wm.WorkTarget, nexts ...func(wm.WorkTarget)) {
		if err := ctx.Err(); err != nil {
			return
		}

		results, err := work(target, map[wm.WorkerName]wm.WorkerConfig{"printer": new(wm.DummyConfig)})
		if err != nil {
			log.Error("work fail: %s", err)
			return
		}

		for _, result := range results {
			if t, ok := result.(*wm.DummyTestTarget); ok && len(t.Remark) < 1024 {
				for _, next := range nexts {
					next(t)
				}
			}
		}
	}, map[wm.WorkerName]wm.WorkerBuilder{
		"printer": PrinterBuilder,
	}, "print")

	mgr.Serve(mgr.ListStep()...)

	task := wm.NewTask(context.Background())
	mgr.AddTask(task)

	task.Cancel()

	// err := mgr.Recv("print", &wm.DummyTestTarget{DummyTarget: wm.DummyTarget{TaskToken: task.Token()}, Step: "print"})
	err := mgr.Recv("print", &wm.DummyTestTarget{DummyTarget: wm.DummyTarget{TaskToken: "token"}, Step: "print"})
	if err != nil {
		fmt.Printf("send target fail: %s", err)
	}

	// time.Sleep(1 * time.Second)
	log.Flush()
	select {}
}
