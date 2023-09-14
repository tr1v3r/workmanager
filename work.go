package workmanager

import (
	"sync"

	"github.com/tr1v3r/pkg/log"
	"github.com/tr1v3r/pkg/pools"
)

// Work workmangager's work
func (wm *WorkerManager) Work(target WorkTarget, configs map[WorkerName]WorkerConfig) (results []WorkTarget, err error) {
	task := wm.GetTask(target.Token())
	if task == nil {
		log.Warn("no such task token %s", target.Token())
		return
	}

	if task.IsCanceled() {
		log.Warn("task(%s) has been canceled\ntask: %+v", task.Token(), target)
		return
	}

	var mu sync.Mutex
	pool := pools.NewPool(defaultPoolSize)
	defer pool.WaitAll()

	for name, conf := range configs {
		if conf == nil || !conf.Active() {
			continue
		}

		select {
		case <-pool.AsyncWait():
		case <-wm.ctx.Done():
			return
		}
		go func(name WorkerName, c WorkerConfig) {
			defer pool.Done()

			builder, ok := wm.workerBuilders[name]
			if !ok {
				log.Error("no such worker builder: %s", name)
				return
			}

			worker := builder(task.Context(), c.Args())
			if worker == nil {
				log.Error("worker builder return nil")
				return
			}

			res, err := wm.work(worker, target)
			if err != nil {
				log.Warn("worker meets error: %s", err)
				return
			}
			if res == nil {
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

	worker.BeforeWork()
	res, err = worker.Work(arg)
	worker.AfterWork()

	select {
	case <-worker.GetContext().Done():
		_ = worker.Terminate()
	case <-worker.Finished():
	}

	return
}
