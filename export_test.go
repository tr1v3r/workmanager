package workmanager_test

import (
	"context"
	"fmt"
	"testing"
	"time"

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
