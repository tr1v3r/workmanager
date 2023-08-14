package workmanager_test

import (
	"context"
	"fmt"
	"time"

	wm "github.com/tr1v3r/workmanager"
)

func ExampleWorkerManager_newInstance() {
	mgr := wm.NewWorkerManager(context.Background())

	mgr.RegisterWorker(wm.DummyWorkerA, wm.DummyBuilder)
	mgr.RegisterWorker(wm.DummyWorkerB, wm.DummyBuilder)

	mgr.RegisterStep(wm.StepA, wm.DummyStepRunner, wm.StepB)
	mgr.RegisterStep(wm.StepB, wm.DummyStepRunner)

	mgr.RegisterBeforeCallbacks(wm.StepA, func(ctx context.Context, t ...wm.WorkTarget) []wm.WorkTarget {
		fmt.Printf("step %s before callback got target: %+v\n", wm.StepA, t[0])
		return t
	})
	mgr.RegisterAfterCallbacks(wm.StepA, func(ctx context.Context, t ...wm.WorkTarget) []wm.WorkTarget {
		fmt.Printf("step %s after callback got target: %+v\n", wm.StepA, t[0])
		return t
	})

	mgr.SetPipe(wm.StepA, wm.PipeChSize(8))

	mgr.Serve()

	task := wm.NewTask(context.Background())
	task.(*wm.Task).TaskToken = "example_token_123"
	mgr.AddTask(task)

	err := mgr.Recv(wm.StepA, &wm.DummyTestTarget{DummyTarget: wm.DummyTarget{TaskToken: task.Token()}, Step: wm.StepA})
	if err != nil {
		fmt.Printf("send target fail: %s", err)
	}

	for c := time.Tick(100 * time.Millisecond); !task.IsFinished(); <-c {
	}

	resultTask := mgr.GetTask(task.Token()).(*wm.Task)
	fmt.Printf("got task: { token: %s, finished: %t }", resultTask.TaskToken, resultTask.Finished)

	// Output:
	// step step_a before callback got target: &{DummyTarget:{TaskToken:example_token_123} Step:step_a}
	// step step_a after callback got target: &{DummyTarget:{TaskToken:example_token_123} Step:step_b}
	// got result: &{DummyTarget:{TaskToken:example_token_123} Step:step_b}
	// got result: &{DummyTarget:{TaskToken:example_token_123} Step:step_a}
	// got task: { token: example_token_123, finished: true }
}

func ExampleWorkerManager_singleton() {
	wm.RegisterWorker(wm.DummyWorkerA, wm.DummyBuilder)
	wm.RegisterWorker(wm.DummyWorkerB, wm.DummyBuilder)

	wm.RegisterStep(wm.StepA, wm.DummyStepRunner, wm.StepB)
	wm.RegisterStep(wm.StepB, wm.DummyStepRunner)

	wm.RegisterBeforeCallbacks(wm.StepA, func(ctx context.Context, t ...wm.WorkTarget) []wm.WorkTarget {
		fmt.Printf("step %s before callback got target: %+v\n", wm.StepA, t[0])
		return t
	})
	wm.RegisterAfterCallbacks(wm.StepA, func(ctx context.Context, t ...wm.WorkTarget) []wm.WorkTarget {
		fmt.Printf("step %s after callback got target: %+v\n", wm.StepA, t[0])
		return t
	})

	wm.SetPipe(wm.StepA, wm.PipeChSize(8))

	wm.Serve()

	task := wm.NewTask(context.Background())
	task.(*wm.Task).TaskToken = "example_token_123"
	wm.AddTask(task)

	err := wm.Recv(wm.StepA, &wm.DummyTestTarget{DummyTarget: wm.DummyTarget{TaskToken: task.Token()}, Step: wm.StepA})
	if err != nil {
		fmt.Printf("send target fail: %s", err)
	}

	for c := time.Tick(100 * time.Millisecond); !task.IsFinished(); <-c {
	}

	resultTask := wm.GetTask(task.Token()).(*wm.Task)
	fmt.Printf("got task: { token: %s, finished: %t }", resultTask.TaskToken, resultTask.Finished)

	// Output:
	// step step_a before callback got target: &{DummyTarget:{TaskToken:example_token_123} Step:step_a}
	// step step_a after callback got target: &{DummyTarget:{TaskToken:example_token_123} Step:step_b}
	// got result: &{DummyTarget:{TaskToken:example_token_123} Step:step_b}
	// got result: &{DummyTarget:{TaskToken:example_token_123} Step:step_a}
	// got task: { token: example_token_123, finished: true }
}
