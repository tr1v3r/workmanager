package workmanager

import (
	"github.com/riverchu/pkg/log"
	"github.com/riverchu/pkg/pools"
)

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

	pool := pools.NewPool(defaultPoolSize * 4)
	defer pool.WaitAll()

	for name, conf := range configs {
		if !conf.Active() {
			continue
		}

		select {
		case <-pool.AsyncWait():
		case <-wm.ctx.Done():
			return
		}
		go func(name WorkerName, c WorkerConfig) {
			defer pool.Done()

			worker := wm.workerBuilders[name](task.Context(), name, c.Args())

			res, err := wm.work(worker, target)
			if err != nil {
				log.Warn("worker meets error: %s", err)
				return
			}
			if res == nil {
				return
			}
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
