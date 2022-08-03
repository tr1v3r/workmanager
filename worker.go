package workmanager

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/time/rate"

	"github.com/riverchu/pkg/log"
)

var workerMgr = NewWorkerManager(context.Background())

// NewWorkerManager ...
func NewWorkerManager(ctx context.Context, opts ...func(*WorkerManager) *WorkerManager) (mgr *WorkerManager) { // nolint
	defer func() {
		for _, opt := range opts {
			mgr = opt(mgr)
		}
	}()
	return &WorkerManager{
		ctx:     ctx,
		pipeMgr: NewPipeManager(ctx),
		taskMgr: NewTaskManager(ctx),
		poolMgr: NewPoolManager(ctx),
		limiter: rate.NewLimiter(5, 100),

		mu:               new(sync.RWMutex),
		workerBuilders:   make(map[WorkerName]WorkerBuilder, 8),
		stepRunners:      make(map[WorkStep]func(*WorkerManager) error, 8),
		resultProcessors: make(map[WorkStep]StepProcessor, 8),
	}
}

type WorkerManager struct {
	ctx context.Context

	pipeMgr *pipeManager
	taskMgr *taskManager
	poolMgr *poolManager
	limiter *rate.Limiter

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

func (wm *WorkerManager) StartStep(step WorkStep, opts ...StepOption) {
	if wm.pipeMgr.Has(step) { // 存在则不需处理
		return
	}
	wm.SetStep(step, opts...)
}

func (wm *WorkerManager) SetStep(step WorkStep, opts ...StepOption) {
	wm.poolMgr.Add(step, 0)
	wm.pipeMgr.SetStep(step, opts...)
}

func (wm *WorkerManager) RemoveStep(steps ...WorkStep) {
	wm.pipeMgr.Remove(steps...)
	wm.poolMgr.Remove(steps...)
}

func (wm *WorkerManager) SetLimit(limit rate.Limit) { wm.limiter = rate.NewLimiter(limit, 100) }

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
		for ch := wm.pipeMgr.GetReadChan(from); ch != nil; _ = wm.limiter.Wait(wm.ctx) {
			select {
			case <-wm.ctx.Done():
				log.Info("step %s runner stopped", from)
				return wm.ctx.Err()
			case target := <-ch:
				wm.run(from, func() {
					defer wm.taskMgr.Done(target.Token())
					runner(wm.Work, target,
						wrap(wm.taskMgr.Get(target.Token()).Start, wm.pipeMgr.GetWriteChans(to...))...,
					)
				})
			}
		}
		return nil
	}
	wm.resultProcessors[from] = processor
}

func wrap(start func(), chs []chan<- WorkTarget) (recvs []func(WorkTarget)) {
	for _, ch := range chs {
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
	pool := wm.poolMgr.Get(step)
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
