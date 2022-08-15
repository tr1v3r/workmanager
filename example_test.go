package workmanager_test

import (
	"context"
	"fmt"
	"time"

	wm "github.com/riverchu/workmanager"
)

func ExampleWorkerManager_NewInstance() {
	fmt.Println("hello world")
	// Output: hello world
}

func ExampleWorkerManager_Singleton() {
	wm.RegisterWorker(DummyWorkerA, dummyBuilder)
	wm.RegisterWorker(DummyWorkerB, dummyBuilder)

	wm.RegisterStep(StepA, dummyStepRunner, dummyStepProcessor, StepB)
	wm.RegisterStep(StepB, dummyStepRunner, dummyStepProcessor)

	wm.Serve()

	task := wm.NewTask(context.Background())
	task.(*wm.Task).TaskToken = "example_token_123"
	wm.AddTask(task)

	err := wm.Recv(StepA, &dummyTarget{DummyTarget: wm.DummyTarget{TaskToken: task.Token()}, step: StepA})
	if err != nil {
		fmt.Printf("send target fail: %s", err)
	}

	for c := time.Tick(100 * time.Millisecond); !task.IsFinished(); <-c {
	}

	resultTask := wm.GetTask(task.Token()).(*wm.Task)
	fmt.Printf("got task: { token: %s, finished: %t }", resultTask.TaskToken, resultTask.Finished)

	// Output:
	// got result: &{DummyTarget:{TaskToken:example_token_123} step:step_b}
	// got result: &{DummyTarget:{TaskToken:example_token_123} step:step_a}
	// got task: { token: example_token_123, finished: true }
}