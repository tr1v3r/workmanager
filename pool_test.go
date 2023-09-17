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
	poolMgr := NewPoolController(context.Background(), poolStepA, poolStepB)

	steps := poolMgr.poolSteps()
	if len(steps) != 2 || !ContainsStep(poolStepA, steps...) || !ContainsStep(poolStepB, steps...) {
		t.Errorf("pool manager init fail: expect steps: %+v, got: %+v", []WorkStep{poolStepA, poolStepB}, steps)
	}
	t.Log("init ok")

	if poolMgr.getPool(poolStepA).Size() != defaultPoolSize || poolMgr.getPool(poolStepB).Size() != defaultPoolSize {
		t.Errorf("pool size error: expect %d, got: %d", defaultPoolSize, poolMgr.getPool(poolStepA).Size())
	}
	t.Log("pool size ok")

	newSize := 99
	poolMgr.SetPool(newSize, poolStepA, poolStepC)
	if poolMgr.getPool(poolStepA).Size() != newSize ||
		poolMgr.getPool(poolStepB).Size() != defaultPoolSize ||
		poolMgr.getPool(poolStepC).Size() != newSize {
		t.Errorf("set pool size fail: expect %d, got %s:%d\t%s:%d\t%s:%d",
			newSize,
			poolStepA, poolMgr.getPool(poolStepA).Size(),
			poolStepB, poolMgr.getPool(poolStepB).Size(),
			poolStepC, poolMgr.getPool(poolStepC).Size(),
		)
	}
	t.Log("set pool ok")

	for _, step := range []WorkStep{poolStepA, poolStepB, poolStepC} {
		dataCh := make(chan struct{}, 999)
		pool := poolMgr.getPool(step)

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

	poolMgr.RemovePool(poolStepA)
	if poolMgr.getPool(poolStepA) != poolMgr.defaultPool {
		t.Errorf("delete step fail: delete poolStepA, got: %+v", poolMgr.poolSteps())
	}
	t.Log("delete pool ok")
}
