package workmanager

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestPipeController_recver(t *testing.T) {
	task := NewTask(context.Background())
	task.(*Task).taskToken = "example_token_123"

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
	t.Logf("got task: { token: %s, finished: %t }\n", resultTask.Token(), resultTask.IsFinished())
}

func TestPipeController_mitm(t *testing.T) {
	mgr := NewPipeController(context.TODO(), StepA, StepB)

	// raw pipe

	// read all data for step A and print to check mitm works or not
	recv := mgr.getRecvChan(StepA)

	const token = "task token"
	mgr.getSendChan(StepA) <- &DummyTestTarget{DummyTarget: DummyTarget{TaskToken: token}, Step: StepA}

	if data := <-recv; data.Token() != token {
		t.Errorf("got data for stepA fail: expect task token: %s, got: %s", token, data.Token())
	}

	// mitmed pipe

	// return origin send ch
	newSend := make(chan WorkTarget, 256)
	originSend := mgr.MITMSendChan(StepA, newSend)

	mgr.getSendChan(StepA) <- &DummyTestTarget{DummyTarget: DummyTarget{TaskToken: token}, Step: StepA}

	const replacedToken = "replaced task token"
	select {
	case data := <-newSend:
		data.(*DummyTestTarget).DummyTarget.TaskToken = replacedToken
		originSend <- data // send data to origin ch after processed
	}

	if data := <-recv; data.Token() != replacedToken {
		t.Errorf("got mitmed data for stepA fail: expect task token: %s, got: %s", replacedToken, data.Token())
	}
}
