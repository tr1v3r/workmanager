package workmanager

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestPipeManager_recver(t *testing.T) {
	mgr := NewWorkerManager(context.Background())

	mgr.RegisterWorker(DummyWorkerA, DummyBuilder)
	mgr.RegisterWorker(DummyWorkerB, DummyBuilder)

	mgr.RegisterStep(StepA, DummyStepRunner, DummyStepProcessor, StepB)
	mgr.RegisterStep(StepB, DummyStepRunner, DummyStepProcessor)

	mgr.SetPipe(StepB, StepRecver(func(target WorkTarget) {
		t.Logf("%s got target %+v, resending...", StepB, target)
	}))

	mgr.Serve()

	task := NewTask(context.Background())
	task.(*Task).TaskToken = "example_token_123"
	mgr.AddTask(task)

	err := mgr.Recv(StepA, &DummyTestTarget{DummyTarget: DummyTarget{TaskToken: task.Token()}, Step: StepA})
	if err != nil {
		fmt.Printf("send target fail: %s", err)
	}

	for c := time.Tick(100 * time.Millisecond); !task.IsFinished(); <-c {
		fmt.Printf("task: %+v", task)
	}

	resultTask := mgr.GetTask(task.Token()).(*Task)
	fmt.Printf("got task: { token: %s, finished: %t }", resultTask.TaskToken, resultTask.Finished)
}
