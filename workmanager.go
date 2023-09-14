package workmanager

import (
	"context"
	"fmt"
	"sync"

	"github.com/tr1v3r/pkg/log"
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

		taskManager:     NewTaskManager(ctx),
		pipeManager:     NewPipeManager(ctx),
		poolManager:     NewPoolManager(ctx),
		limitManager:    NewLimitManager(ctx),
		callbackManager: newCallbackManager(),

		mu:             new(sync.RWMutex),
		workerBuilders: make(map[WorkerName]WorkerBuilder, 8),
		stepRunners:    make(map[WorkStep]func(*WorkerManager) error, 8),
	}
}

// WorkerManager worker manager
type WorkerManager struct {
	ctx context.Context

	cacher Cacher

	*taskManager
	*pipeManager
	*poolManager
	*limitManager
	*callbackManager

	mu             *sync.RWMutex
	workerBuilders map[WorkerName]WorkerBuilder
	stepRunners    map[WorkStep]func(*WorkerManager) error
}

// WithContext set context
func (wm WorkerManager) WithContext(ctx context.Context) *WorkerManager {
	wm.ctx = ctx
	wm.taskManager = wm.taskManager.WithContext(ctx)
	return &wm
}

// SetCacher set cacher
func (wm *WorkerManager) SetCacher(c Cacher) { wm.cacher = c }

// StartStep  start step
func (wm *WorkerManager) StartStep(step WorkStep, opts ...PipeOption) {
	if wm.HasPipe(step) { // 存在则不需处理
		return
	}
	wm.InitStep(step, opts...)
}

// InitStep initialize step
func (wm *WorkerManager) InitStep(step WorkStep, opts ...PipeOption) {
	if wm.getPool(step) == wm.defaultPool {
		wm.SetPool(0, step)
	}
	if wm.getLimiter(step) == wm.defaultLimiter {
		wm.SetLimiter(0, 0, step)
	}
	wm.SetPipe(step, opts...)
}

// RemoveStep remove steps
func (wm *WorkerManager) RemoveStep(steps ...WorkStep) {
	wm.RemovePipe(steps...)
	wm.RemovePool(steps...)
}

// ListStep list steps
func (wm *WorkerManager) ListStep() (steps []WorkStep) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	for step := range wm.stepRunners {
		steps = append(steps, step)
	}
	return steps
}

// Register register step, runner and workers
func (wm *WorkerManager) Register(
	current WorkStep,
	runner StepRunner,
	workers map[WorkerName]WorkerBuilder,
	nexts ...WorkStep,
) {
	for name, builder := range workers {
		wm.RegisterWorker(name, builder)
	}
	wm.RegisterStep(current, runner, nexts...)
}

// RegisterWorker register worker
func (wm *WorkerManager) RegisterWorker(
	name WorkerName,
	builder WorkerBuilder,
) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.workerBuilders[name] = builder
}

// RegisterStep register step and its runner
// step runner determines how the step works, which workers to call, what configurations to use for each,
// and what the next step is
func (wm *WorkerManager) RegisterStep(current WorkStep, runner StepRunner, nexts ...WorkStep) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.stepRunners[current] = func(wm *WorkerManager) error {
		defer catchPanic("%s step runner panic", current)

		for ch := wm.GetRecvChan(current); ch != nil; _ = wm.getLimiter(current).Wait(wm.ctx) {
			select {
			case <-wm.ctx.Done():
				log.Info("step %s runner stopped", current)
				return wm.ctx.Err()
			case target := <-ch:
				wm.run(current, func() {
					defer catchPanic("%s step work panic", current)

					task := wm.GetTask(target.Token())
					if wm.cacher != nil && !wm.cacher.Allow(target) {
						return
					}

					callbacks := wm.GetCallbacks(current)

					runner(
						task.Context(),
						wrapWork(task.Context(), callbacks.BeforeWork(), wm.Work, callbacks.AfterWork()),
						target,
						wrapChan(wm.GetSendChans(nexts...))...,
					)
				})
			}
		}
		return nil
	}
}

func wrapWork(ctx context.Context, before []StepCallback, work Work, after []StepCallback) Work {
	return func(target WorkTarget, configs map[WorkerName]WorkerConfig) ([]WorkTarget, error) {
		for _, call := range before {
			if results := call(ctx, target); len(results) > 0 {
				target = results[0]
			} else {
				target = nil
			}
		}
		results, err := work(target, configs)
		if err != nil {
			return nil, err
		}
		for _, res := range results {
			res.SetToken(target.Token())
		}
		for _, call := range after {
			results = call(ctx, results...)
		}
		return results, nil
	}
}

func wrapChan(chs []chan<- WorkTarget) (recvs []func(WorkTarget)) {
	for _, ch := range chs {
		ch := ch
		recvs = append(recvs, func(target WorkTarget) { ch <- target })
	}
	return recvs
}

func (wm *WorkerManager) run(step WorkStep, runner func()) {
	pool := wm.getPool(step)

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

// ============ callbacks ============

// RegisterBeforeCallbacks register before callback funcs
func (wm *WorkerManager) RegisterBeforeCallbacks(step WorkStep, callbacks ...StepCallback) {
	wm.GetCallbacks(step).RegisterBefore(callbacks...)
}

// RegisterAfterCallbacks register after callbacks
func (wm *WorkerManager) RegisterAfterCallbacks(step WorkStep, callbacks ...StepCallback) {
	wm.GetCallbacks(step).RegisterAfter(callbacks...)
}

// Serve start serve with specifid steps, do nothing when steps == nil
func (wm *WorkerManager) Serve(steps ...WorkStep) {
	log.Info("starting worker routine...")

	wm.mu.RLock()
	defer wm.mu.RUnlock()
	for _, step := range steps {
		wm.StartStep(step)
		go wm.stepRunners[step](wm) // nolint
	}
}

// Recv recv func
func (wm *WorkerManager) Recv(step WorkStep, target WorkTarget) error {
	if ch := wm.GetSendChan(step); ch != nil {
		ch <- target
		return nil
	}
	return fmt.Errorf("%s channel not found", step)
}

// RecvFrom recv from chan
func (wm *WorkerManager) RecvFrom(step WorkStep, recv <-chan WorkTarget) error {
	go func() {
		for {
			select {
			case <-wm.ctx.Done():
				return
			case target, ok := <-recv:
				if !ok {
					return
				}
				_ = wm.Recv(step, target)
			}
		}
	}()
	return nil
}
