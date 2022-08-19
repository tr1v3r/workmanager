package workmanager

import (
	"context"
	"testing"
	"time"
)

func TestPipeManager_recver(t *testing.T) {
	mgr := NewWorkerManager(context.Background())

	mgr.RegisterWorker(DummyWorkerA, DummyBuilder)
	mgr.RegisterWorker(DummyWorkerB, DummyBuilder)

	var endGame bool
	mgr.RegisterStep(StepA, DummyStepRunner, DummyStepProcessor, StepB)
	mgr.RegisterStep(StepB, func(_ Work, target WorkTarget, _ ...func(WorkTarget)) {
		t.Logf("%s got target %+v, transfering...", StepB, target)
		endGame = true
	}, nil)

	mgr.Serve()

	task := NewTask(context.Background())
	task.(*Task).TaskToken = "example_token_123"
	mgr.AddTask(task)

	err := mgr.Recv(StepA, &DummyTestTarget{DummyTarget: DummyTarget{TaskToken: task.Token()}, Step: StepA})
	if err != nil {
		t.Logf("send target fail: %s\n", err)
	}

	for c := time.Tick(100 * time.Millisecond); !endGame; <-c {
		t.Logf("task: %+v\n", task)
	}

	resultTask := mgr.GetTask(task.Token()).(*Task)
	t.Logf("got task: { token: %s, finished: %t }\n", resultTask.TaskToken, resultTask.Finished)
}
