package workmanager

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestPipeController_recver(t *testing.T) {
	task := NewTask(context.Background())
	task.(*Task).TaskToken = "example_token_123"

	mgr := NewWorkerManager(context.Background())

	mgr.RegisterWorker(DummyWorkerA, DummyBuilder(DummyWorkerA))
	mgr.RegisterWorker(DummyWorkerB, DummyBuilder(DummyWorkerB))

	mgr.RegisterStep(StepA, DummyStepRunner, StepB)
	mgr.RegisterStep(StepB, TransferRunner(func(_ context.Context, target WorkTarget) {
		fmt.Printf("%s got target %+v, transfering...\n", StepB, target)
		task.Finish()
	}))

	mgr.Serve(StepA, StepB)

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

func TestPipeController_mitm(t *testing.T) {
	mgr := NewPipeController(nil, StepA, StepB)

	// read all data for step A and print to check mitm works or not
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

	newSendChan := make(chan WorkTarget, 256)
	// send := mgr.GetSendChan(StepA)
	// mgr.SetSendChan(StepA, newSendChan)

	// return origin send ch
	send := mgr.MITMSendChan(StepA, newSendChan)

	mgr.GetSendChan(StepA) <- &DummyTestTarget{DummyTarget: DummyTarget{TaskToken: "Raw Target 2"}, Step: StepA}

	select {
	case data := <-newSendChan:
		data.(*DummyTestTarget).DummyTarget.TaskToken = "Converted Target"
		send <- data // send data to origin ch after processed
	}

	time.Sleep(time.Second)
}
