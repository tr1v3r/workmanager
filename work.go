package workmanager

import (
	"fmt"
	"sync"

	"github.com/tr1v3r/pkg/log"
	"github.com/tr1v3r/pkg/pools"
)

// Work workmangager's work
// target: work target
// config: workers' config, key is worker's name, value is worker's config, build a new worker when config is not nil and active
func (wm *WorkerManager) Work(target WorkTarget, configs map[WorkerName]WorkerConfig) (results []WorkTarget, err error) {
	task := wm.GetTask(target.Token())
	if task == nil {
		log.Warn("task not found: token %s", target.Token())
		return
	}

	if task.IsCanceled() {
		log.Warn("task(%s) has been canceled\ntask: %+v", task.Token(), target)
		return
	}

	var mu sync.Mutex
	pool := pools.NewPool(defaultPoolSize)
	defer pool.WaitAll()

	// build workers and work 1 by 1
	for name, conf := range configs {
		if conf == nil || !conf.Active() {
			continue
		}

		// check gorountine pool and wm's context
		select {
		case <-pool.AsyncWait():
		case <-wm.ctx.Done():
			return // stop Work when wm's context done
		}

		// build worker and work
		go func(name WorkerName, c WorkerConfig) {
			defer pool.Done()

			build, ok := wm.workerBuilders[name]
			if !ok {
				log.Error("worker builder not found: %s", name)
				return
			}

			// build worker with wm's context and config args
			worker := build(task.Context(), c.Args())
			if worker == nil {
				log.Error("bulid worker fail: build %s got nil", name)
				return
			}

			// work
			res, err := wm.work(worker, target)
			if err != nil {
				log.Warn("%s work fail: %s", name, err)
				return
			}
			if res == nil {
				log.Debug("%s work on %s got nil result", name, target.Key())
				return
			}
			mu.Lock()
			defer mu.Unlock()
			results = append(results, res...)
		}(name, conf)
	}

	return
}

func (wm *WorkerManager) work(worker Worker, arg WorkTarget) (res []WorkTarget, err error) {
	defer catchPanic("work panic")

	// call worker's Work
	res, err = worker.Work(arg)
	if err != nil {
		return nil, fmt.Errorf("work fail: %w", err)
	}
	return res, nil
}
