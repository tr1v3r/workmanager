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
	limitCtr := NewLimitController(context.Background(), limitStepA, limitStepB)

	if steps := limitCtr.limitSteps(); len(steps) != 2 || !ContainsStep(limitStepA, steps...) || !ContainsStep(limitStepB, steps...) {
		t.Errorf("limit manager init fail: expect steps: %+v, got: %+v", []WorkStep{limitStepA, limitStepB}, steps)
	}
	t.Log("init ok")

	if limitCtr.getLimiter(limitStepA).Limit() != defaultLimit || limitCtr.getLimiter(limitStepB).Limit() != defaultLimit {
		t.Errorf("limit size error: expect %f, got: %f", limitCtr.getLimiter(limitStepA).Limit(), defaultLimit)
	}
	t.Log("limit size ok")

	var newLimit rate.Limit = 9
	limitCtr.SetLimit(newLimit, 1, limitStepA, limitStepC)
	if limitCtr.getLimiter(limitStepA).Limit() != newLimit ||
		limitCtr.getLimiter(limitStepB).Limit() != defaultLimit ||
		limitCtr.getLimiter(limitStepC).Limit() != newLimit {
		t.Errorf("set limit size fail: expect %f, got %s:%f\t%s:%f\t%s:%f",
			newLimit,
			limitStepA, limitCtr.getLimiter(limitStepA).Limit(),
			limitStepB, limitCtr.getLimiter(limitStepB).Limit(),
			limitStepC, limitCtr.getLimiter(limitStepC).Limit(),
		)
	}
	t.Log("set limit ok")

	for _, step := range []WorkStep{limitStepA, limitStepB, limitStepC} {
		dataCh := make(chan struct{}, 999)
		limiter := limitCtr.getLimiter(step)

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

	limitCtr.DelLimiter(limitStepA)
	if limitCtr.getLimiter(limitStepA) != limitCtr.defaultLimiter {
		t.Errorf("delete step fail: delete limitStepA, got: %+v", limitCtr.limitSteps())
	}
	t.Log("delete limiter ok")
}
