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

	mgr.RegisterStep(StepA, DummyStepRunner, StepB)
	mgr.RegisterStep(StepB, TransferRunner(func(target WorkTarget) {
		fmt.Printf("%s got target %+v, transfering...", StepB, target)
	}))

	mgr.Serve()

	task := NewTask(context.Background())
	task.(*Task).TaskToken = "example_token_123"
	mgr.AddTask(task)

	err := mgr.Recv(StepA, &DummyTestTarget{DummyTarget: DummyTarget{TaskToken: task.Token()}, Step: StepA})
	if err != nil {
		t.Logf("send target fail: %s\n", err)
	}

	for c := time.Tick(100 * time.Millisecond); !task.IsFinished(); <-c {
		t.Logf("task: %+v\n", task)
	}

	resultTask := mgr.GetTask(task.Token()).(*Task)
	t.Logf("got task: { token: %s, finished: %t }\n", resultTask.TaskToken, resultTask.Finished)
}

func TestPipeManager_mitm(t *testing.T) {
	mgr := NewPipeManager(nil, StepA, StepB)

	// read
	recv := mgr.GetRecvChan(StepA)
	go func() {
		for {
			select {
			case data := <-recv:
				t.Logf("got data: %s", data.Token())
			}
		}
	}()

	mgr.GetSendChan(StepA) <- &DummyTestTarget{DummyTarget: DummyTarget{TaskToken: "Raw Target 1"}, Step: StepA}

	send := mgr.GetSendChan(StepA)
	newSendChan := make(chan WorkTarget, 256)
	mgr.SetSendChan(StepA, newSendChan)
	mgr.GetSendChan(StepA) <- &DummyTestTarget{DummyTarget: DummyTarget{TaskToken: "Raw Target 2"}, Step: StepA}

	select {
	case data := <-newSendChan:
		data.(*DummyTestTarget).DummyTarget.TaskToken = "Converted Target"
		send <- data
	}
	time.Sleep(time.Second)
}
