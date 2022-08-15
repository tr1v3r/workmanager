package workmanager_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	wm "github.com/riverchu/workmanager"
)

func Example_NewInstance() {
	fmt.Println("hello world")
	// Output: hello world
}

func Example_Singleton() {
	fmt.Println("hello world")
	// Output: hello world
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

	for c := time.Tick(100 * time.Millisecond); !task.IsFinished(); <-c {
	}

	t.Logf("task %+v", wm.GetTask(task.Token()))
}
