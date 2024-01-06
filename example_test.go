package workmanager_test

import (
	"context"
	"fmt"
	"time"

	wm "github.com/tr1v3r/workmanager"
)

func ExampleWorkerManager_newInstance() {
	mgr := wm.NewWorkerManager(context.Background())

	// register worker by workerbuilder with name
	mgr.RegisterWorker(wm.DummyWorkerA, wm.DummyBuilder(wm.DummyWorkerA))
	mgr.RegisterWorker(wm.DummyWorkerB, wm.DummyBuilder(wm.DummyWorkerB))

	// register step, specify from which step to which step
	mgr.RegisterStep(wm.StepA, wm.DummyStepRunner, wm.StepB)
	mgr.RegisterStep(wm.StepB, wm.DummyStepRunner)

	// register hooks
	mgr.RegisterBeforeCallbacks(wm.StepA, func(_ context.Context, t ...wm.WorkTarget) []wm.WorkTarget {
		fmt.Printf("[%s] before callback got target: %+v\n", wm.StepA, t[0])
		return t
	})
	mgr.RegisterAfterCallbacks(wm.StepA, func(_ context.Context, t ...wm.WorkTarget) []wm.WorkTarget {
		fmt.Printf("[%s] after callback got target: %+v\n", wm.StepA, t[0])
		return t
	})
	mgr.RegisterAfterCallbacks(wm.StepB, func(_ context.Context, t ...wm.WorkTarget) []wm.WorkTarget {
		mgr.FinishTask(t[0].Token())
		return t
	})

	mgr.SetPipe(wm.StepA, wm.PipeChSize(8))

	// start serve
	mgr.Serve(wm.StepA, wm.StepB)

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
	fmt.Printf("task final status: { token: %s, finished: %t }", resultTask.TaskToken, resultTask.Finished)

	// Output:
	// [step_a] before callback got target: &{DummyTarget:{TaskToken:example_token_123} Step:step_a Remark: Count:0}
	// [step_a] after callback got target: &{DummyTarget:{TaskToken:example_token_123} Step:step_b Remark: Count:0}
	// [step_a] got result: &{DummyTarget:{TaskToken:example_token_123} Step:step_b Remark: Count:0}
	// [step_b] got result: &{DummyTarget:{TaskToken:example_token_123} Step:step_a Remark: Count:0}
	// task final status: { token: example_token_123, finished: true }
}

func ExampleWorkerManager_singleton() {
	wm.RegisterWorker(wm.DummyWorkerA, wm.DummyBuilder(wm.DummyWorkerA))
	wm.RegisterWorker(wm.DummyWorkerB, wm.DummyBuilder(wm.DummyWorkerB))

	wm.RegisterStep(wm.StepA, wm.DummyStepRunner, wm.StepB)
	wm.RegisterStep(wm.StepB, wm.DummyStepRunner)

	wm.RegisterBeforeCallbacks(wm.StepA, func(_ context.Context, t ...wm.WorkTarget) []wm.WorkTarget {
		fmt.Printf("[%s] before callback got target: %+v\n", wm.StepA, t[0])
		return t
	})
	wm.RegisterAfterCallbacks(wm.StepA, func(_ context.Context, t ...wm.WorkTarget) []wm.WorkTarget {
		fmt.Printf("[%s] after callback got target: %+v\n", wm.StepA, t[0])
		return t
	})
	wm.RegisterAfterCallbacks(wm.StepB, func(_ context.Context, t ...wm.WorkTarget) []wm.WorkTarget {
		wm.FinishTask(t[0].Token())
		return t
	})

	wm.SetPipe(wm.StepA, wm.PipeChSize(8))

	wm.Serve(wm.StepA, wm.StepB)

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
	fmt.Printf("task final status: { token: %s, finished: %t }", resultTask.TaskToken, resultTask.Finished)

	// Output:
	// [step_a] before callback got target: &{DummyTarget:{TaskToken:example_token_123} Step:step_a Remark: Count:0}
	// [step_a] after callback got target: &{DummyTarget:{TaskToken:example_token_123} Step:step_b Remark: Count:0}
	// [step_a] got result: &{DummyTarget:{TaskToken:example_token_123} Step:step_b Remark: Count:0}
	// [step_b] got result: &{DummyTarget:{TaskToken:example_token_123} Step:step_a Remark: Count:0}
	// task final status: { token: example_token_123, finished: true }
}
