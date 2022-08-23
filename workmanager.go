package workmanager

import (
	"context"
	"fmt"
	"sync"

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

func (wm WorkerManager) WithContext(ctx context.Context) *WorkerManager {
	wm.ctx = ctx
	wm.taskManager = wm.taskManager.WithContext(ctx)
	return &wm
}

func (wm *WorkerManager) SetCacher(c Cacher) { wm.cacher = c }

func (wm *WorkerManager) StartStep(step WorkStep, opts ...PipeOption) {
	if wm.HasPipe(step) { // 存在则不需处理
		return
	}
	wm.SetStep(step, opts...)
}

func (wm *WorkerManager) SetStep(step WorkStep, opts ...PipeOption) {
	wm.SetPool(0, step)
	wm.SetPipe(step, opts...)
}

func (wm *WorkerManager) RemoveStep(steps ...WorkStep) {
	wm.DelPipe(steps...)
	wm.DelPool(steps...)
}

func (wm *WorkerManager) ListSteps() (steps []WorkStep) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	for step := range wm.stepRunners {
		steps = append(steps, step)
	}
	return steps
}

func (wm *WorkerManager) Register(
	from WorkStep,
	runner StepRunner,
	workers map[WorkerName]WorkerBuilder,
	to ...WorkStep,
) {
	for name, builder := range workers {
		wm.RegisterWorker(name, builder)
	}
	wm.RegisterStep(from, runner, to...)
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
	to ...WorkStep,
) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.stepRunners[from] = func(wm *WorkerManager) error {
		defer catchPanic("%s step runner panic", from)

		for ch := wm.GetRecvChan(from); ch != nil; _ = wm.GetLimiter(from).Wait(wm.ctx) {
			select {
			case <-wm.ctx.Done():
				log.Info("step %s runner stopped", from)
				return wm.ctx.Err()
			case target := <-ch:
				wm.run(from, func() {
					defer catchPanic("%s step work panic", from)

					task := wm.GetTask(target.Token())
					defer task.Done()
					if wm.cacher != nil && !wm.cacher.Allow(target) {
						return
					}

					callbacks := wm.GetCallbacks(from)

					runner(
						wrapWork(callbacks.BeforeWork(), wm.Work, callbacks.AfterWork()),
						target,
						wrapChan(task.Start, wm.GetSendChans(to...))...,
					)
				})
			}
		}
		return nil
	}
}

func wrapWork(before []StepCallback, work Work, after []StepCallback) Work {
	return func(target WorkTarget, configs map[WorkerName]WorkerConfig) ([]WorkTarget, error) {
		for _, call := range before {
			if results := call(target); len(results) > 0 {
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
			results = call(results...)
		}
		return results, nil
	}
}

func wrapChan(start func() error, chs []chan<- WorkTarget) (recvs []func(WorkTarget)) {
	for _, ch := range chs {
		ch := ch
		recvs = append(recvs, func(target WorkTarget) {
			start()
			ch <- target
		})
	}
	return recvs
}

func (wm *WorkerManager) run(step WorkStep, runner func()) {
	pool := wm.GetPool(step)
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

// ============ callbacks ============
func (wm *WorkerManager) RegisterBeforeCallbacks(step WorkStep, callbacks ...StepCallback) {
	wm.GetCallbacks(step).RegisterBefore(callbacks...)
}
func (wm *WorkerManager) RegisterAfterCallbacks(step WorkStep, callbacks ...StepCallback) {
	wm.GetCallbacks(step).RegisterAfter(callbacks...)
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
	ch := wm.GetSendChan(step)
	if ch == nil {
		return fmt.Errorf("%s channel not found", step)
	}

	if err := wm.TaskStart(target.Token()); err != nil {
		return fmt.Errorf("start task fail: %w", err)
	}

	ch <- target

	return nil
}

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
				wm.Recv(step, target)
			}
		}
	}()
	return nil
}
