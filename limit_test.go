package workmanager

import (
	"context"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

const (
	limitStepA WorkStep = "limit_stepA"
	limitStepB WorkStep = "limit_stepB"
	limitStepC WorkStep = "limit_stepC"
)

func Test_limitController(t *testing.T) {
	limitMgr := NewLimitController(context.Background(), limitStepA, limitStepB)

	steps := limitMgr.limitSteps()
	if len(steps) != 2 || !ContainsStep(limitStepA, steps...) || !ContainsStep(limitStepB, steps...) {
		t.Errorf("limit manager init fail: expect steps: %+v, got: %+v", []WorkStep{limitStepA, limitStepB}, steps)
	}
	t.Log("init ok")

	if limitMgr.getLimiter(limitStepA).Limit() != defaultStepLimit || limitMgr.getLimiter(limitStepB).Limit() != defaultStepLimit {
		t.Errorf("limit size error: expect %f, got: %f", limitMgr.getLimiter(limitStepA).Limit(), defaultStepLimit)
	}
	t.Log("limit size ok")

	var newLimit rate.Limit = 9
	limitMgr.SetLimiter(newLimit, 1, limitStepA, limitStepC)
	if limitMgr.getLimiter(limitStepA).Limit() != newLimit ||
		limitMgr.getLimiter(limitStepB).Limit() != defaultStepLimit ||
		limitMgr.getLimiter(limitStepC).Limit() != newLimit {
		t.Errorf("set limit size fail: expect %f, got %s:%f\t%s:%f\t%s:%f",
			newLimit,
			limitStepA, limitMgr.getLimiter(limitStepA).Limit(),
			limitStepB, limitMgr.getLimiter(limitStepB).Limit(),
			limitStepC, limitMgr.getLimiter(limitStepC).Limit(),
		)
	}
	t.Log("set limit ok")

	for _, step := range []WorkStep{limitStepA, limitStepB, limitStepC} {
		dataCh := make(chan struct{}, 999)
		limiter := limitMgr.getLimiter(step)

		var timeup bool
		for tick := time.Tick(time.Second); !timeup; {
			select {
			case <-tick:
				timeup = true
				break
			default:
				if limiter.Allow() {
					dataCh <- struct{}{}
				}
			}
		}
		if len(dataCh) != int(limiter.Limit())+defaultBurst {
			t.Errorf("%s limit size error: expect %f, got %d", step, limiter.Limit(), len(dataCh))
		}
	}
	t.Log("limiter works ok")

	limitMgr.DelLimiter(limitStepA)
	if limitMgr.getLimiter(limitStepA) != limitMgr.defaultLimiter {
		t.Errorf("delete step fail: delete limitStepA, got: %+v", limitMgr.limitSteps())
	}
	t.Log("delete limiter ok")
}
