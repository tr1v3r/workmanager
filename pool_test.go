package workmanager

import (
	"context"
	"testing"
	"time"
)

const (
	poolStepA WorkStep = "pool_stepA"
	poolStepB WorkStep = "pool_stepB"
	poolStepC WorkStep = "pool_stepC"
)

func Test_poolController(t *testing.T) {
	poolCtr := NewPoolController(context.Background(), poolStepA, poolStepB)

	if steps := poolCtr.poolSteps(); len(steps) != 2 || !ContainsStep(poolStepA, steps...) || !ContainsStep(poolStepB, steps...) {
		t.Errorf("pool manager init fail: expect steps: %+v, got: %+v", []WorkStep{poolStepA, poolStepB}, steps)
	}
	t.Log("init ok")

	if poolCtr.getPool(poolStepA).Size() != defaultPoolSize || poolCtr.getPool(poolStepB).Size() != defaultPoolSize {
		t.Errorf("pool size error: expect %d, got: %d", defaultPoolSize, poolCtr.getPool(poolStepA).Size())
	}
	t.Log("pool size ok")

	newSize := 99
	poolCtr.SetPool(newSize, poolStepA, poolStepC)
	if poolCtr.getPool(poolStepA).Size() != newSize ||
		poolCtr.getPool(poolStepB).Size() != defaultPoolSize ||
		poolCtr.getPool(poolStepC).Size() != newSize {
		t.Errorf("set pool size fail: expect %d, got %s:%d\t%s:%d\t%s:%d",
			newSize,
			poolStepA, poolCtr.getPool(poolStepA).Size(),
			poolStepB, poolCtr.getPool(poolStepB).Size(),
			poolStepC, poolCtr.getPool(poolStepC).Size(),
		)
	}
	t.Log("set pool ok")

	for _, step := range []WorkStep{poolStepA, poolStepB, poolStepC} {
		dataCh := make(chan struct{}, 999)
		pool := poolCtr.getPool(step)

		var timeup bool
		for tick := time.Tick(100 * time.Millisecond); !timeup; {
			select {
			case <-tick:
				timeup = true
				break
			default:
				pool.Wait()
				go func() {
					defer pool.Done()
					dataCh <- struct{}{}
					time.Sleep(120 * time.Millisecond)
				}()
			}
		}
		if len(dataCh) != pool.Size() {
			t.Errorf("%s pool size error: expect %d, got %d", step, pool.Size(), len(dataCh))
		}
	}
	t.Log("pool works ok")

	poolCtr.RemovePool(poolStepA)
	if poolCtr.getPool(poolStepA) != poolCtr.defaultPool {
		t.Errorf("delete step fail: delete poolStepA, got: %+v", poolCtr.poolSteps())
	}
	t.Log("delete pool ok")
}
