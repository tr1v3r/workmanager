package workmanager

import (
	"errors"

	"github.com/riverchu/pkg/log"
	"github.com/riverchu/pkg/pools"
)

func (wm *WorkerManager) Work(target WorkTarget, configs map[WorkerName]WorkerConfig) (results []WorkTarget, err error) {
	task := wm.taskMgr.Get(target.Token())
	if task == nil {
		log.Warn("no such task token %s", target.Token())
		return
	}

	if task.IsCanceled() {
		log.Warn("task(%s) has been canceled\ntask: %+v", task.Token(), target)
		return
	}

	pool := pools.NewPool(wm.poolMgr.Size() * 4)
	for name, conf := range configs {
		if !conf.Active() {
			continue
		}

		_ = wm.limiter.Wait(wm.ctx)
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
			results = append(results, res)
		}(name, conf)
	}
	pool.WaitAll()

	return wm.ProcessResult(results...)
}

func (wm *WorkerManager) work(worker Worker, arg WorkTarget) (res WorkTarget, err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("work panic: %v\n%v", err, catchStack())
		}
	}()

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

// ProcessResult resolve result
func (wm *WorkerManager) ProcessResult(results ...WorkTarget) (processedResult []WorkTarget, err error) {
	if len(results) == 0 {
		return nil, nil
	}
	log.Info("resolving %d results", len(results))

	resultMap := make(map[WorkStep][]WorkTarget, 4)
	for _, res := range results {
		if step := res.Step(); step != "" {
			resultMap[step] = append(resultMap[step], res)
		}
	}
	for step, results := range resultMap {
		processor := wm.resultProcessors[step]
		if processor == nil {
			return nil, errors.New("step %s result processor not found")
		}
		result, err := processor(results...)
		if err != nil {
			return nil, err
		}
		processedResult = append(processedResult, result...)
	}
	return processedResult, nil
}
