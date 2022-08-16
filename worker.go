package workmanager

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/time/rate"

	"github.com/riverchu/pkg/log"
)

var defaultWorkerMgr = NewWorkerManager(context.Background())

// NewWorkerManager ...
func NewWorkerManager(ctx context.Context, opts ...func(*WorkerManager) *WorkerManager) (mgr *WorkerManager) { // nolint
	defer func() {
		for _, opt := range opts {
			mgr = opt(mgr)
		}
	}()
	return &WorkerManager{
		ctx: ctx,

		pipeMgr:  NewPipeManager(ctx),
		taskMgr:  NewTaskManager(ctx),
		poolMgr:  NewPoolManager(ctx),
		limitMgr: NewLimitManager(ctx),

		mu:               new(sync.RWMutex),
		workerBuilders:   make(map[WorkerName]WorkerBuilder, 8),
		stepRunners:      make(map[WorkStep]func(*WorkerManager) error, 8),
		resultProcessors: make(map[WorkStep]StepProcessor, 8),
	}
}

type WorkerManager struct {
	ctx context.Context

	cacher Cacher

	pipeMgr  *pipeManager
	taskMgr  *taskManager
	poolMgr  *poolManager
	limitMgr *limitManager

	mu               *sync.RWMutex
	workerBuilders   map[WorkerName]WorkerBuilder
	stepRunners      map[WorkStep]func(*WorkerManager) error
	resultProcessors map[WorkStep]StepProcessor
}

func (wm WorkerManager) WithContext(ctx context.Context) *WorkerManager {
	wm.ctx = ctx
	wm.taskMgr = wm.taskMgr.WithContext(ctx)
	return &wm
}

func (wm *WorkerManager) SetCacher(c Cacher) { wm.cacher = c }

func (wm *WorkerManager) StartStep(step WorkStep, opts ...StepOption) {
	if wm.pipeMgr.Has(step) { // 存在则不需处理
		return
	}
	wm.SetStep(step, opts...)
}

func (wm *WorkerManager) SetStep(step WorkStep, opts ...StepOption) {
	wm.poolMgr.SetPool(0, step)
	wm.pipeMgr.SetStep(step, opts...)
}

func (wm *WorkerManager) RemoveStep(steps ...WorkStep) {
	wm.pipeMgr.Remove(steps...)
	wm.poolMgr.DelPool(steps...)
}

func (wm *WorkerManager) SetLimit(limit rate.Limit, b int) { wm.limitMgr.SetDefaultLimiter(limit, b) }

func (wm *WorkerManager) SetStepLimit(step WorkStep, r rate.Limit, b int) {
	wm.limitMgr.SetLimiter(step, r, b)
}

func (wm *WorkerManager) Register(
	from WorkStep,
	runner StepRunner,
	processor StepProcessor,
	workers map[WorkerName]WorkerBuilder,
	to ...WorkStep,
) {
	for name, builder := range workers {
		wm.RegisterWorker(name, builder)
	}
	wm.RegisterStep(from, runner, processor, to...)
}

func (wm *WorkerManager) RegisterWorker(
	name WorkerName,
	builder WorkerBuilder,
) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.workerBuilders[name] = builder
}

func (wm *WorkerManager) RegisterStep(
	from WorkStep,
	runner StepRunner,
	processor StepProcessor,
	to ...WorkStep,
) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.stepRunners[from] = func(wm *WorkerManager) error {
		defer catchPanic("%s step runner panic", from)

		for ch := wm.pipeMgr.GetReadChan(from); ch != nil; _ = wm.limitMgr.GetLimiter(from).Wait(wm.ctx) {
			select {
			case <-wm.ctx.Done():
				log.Info("step %s runner stopped", from)
				return wm.ctx.Err()
			case target := <-ch:
				wm.run(from, func() {
					defer catchPanic("%s step work panic", from)

					task := wm.taskMgr.Get(target.Token())
					defer task.Done()
					if wm.cacher != nil && !wm.cacher.Allow(target) {
						return
					}
					runner(
						wrapWork(wm.Work, wm.resultProcessors[from]),
						target,
						wrapChan(task.Start, wm.pipeMgr.GetWriteChans(to...))...,
					)
				})
			}
		}
		return nil
	}
	wm.resultProcessors[from] = processor
}

func wrapWork(work Work, processor StepProcessor) Work {
	return func(target WorkTarget, configs map[WorkerName]WorkerConfig) ([]WorkTarget, error) {
		results, err := work(target, configs)
		if err != nil {
			return nil, err
		}
		for _, res := range results {
			res.SetToken(target.Token())
		}
		if processor == nil {
			return results, nil
		}
		return processor(results...)
	}
}

func wrapChan(start func(), chs []chan<- WorkTarget) (recvs []func(WorkTarget)) {
	for _, ch := range chs {
		ch := ch
		recvs = append(recvs, func(target WorkTarget) {
			start()
			ch <- target
		})
	}
	return recvs
}

func (wm *WorkerManager) Serve(steps ...WorkStep) {
	log.Info("starting worker routine...")

	wm.mu.RLock()
	defer wm.mu.RUnlock()
	if len(steps) > 0 {
		for _, step := range steps {
			wm.StartStep(step)
			go wm.stepRunners[step](wm) // nolint
		}
	} else {
		for step, runner := range wm.stepRunners {
			wm.StartStep(step)
			go runner(wm) // nolint
		}
	}
}

func (wm *WorkerManager) Recv(step WorkStep, target WorkTarget) error {
	ch := wm.pipeMgr.GetWriteChan(step)
	if ch == nil {
		return fmt.Errorf("%s channel not found", step)
	}

	wm.taskMgr.Start(target.Token())

	ch <- target

	return nil
}

func (wm *WorkerManager) run(step WorkStep, runner func()) {
	pool := wm.poolMgr.GetPool(step)
	if pool == nil {
		log.Warn("step %s's pool not found, task will not run", step)
		return
	}

	select {
	case <-pool.AsyncWait():
	case <-wm.ctx.Done():
		return
	}

	go func() {
		defer pool.Done()
		runner()
	}()
}

// ==================== Task API ====================

// AddTask add new task
func (wm *WorkerManager) AddTask(task WorkTask) { wm.taskMgr.Add(task) }

// GetTask get task object
func (wm *WorkerManager) GetTask(token string) WorkTask { return wm.taskMgr.Get(token) }

// CancelTask cancel task
func (wm *WorkerManager) CancelTask(token string) error { return wm.taskMgr.Cancel(token) }
