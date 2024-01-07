package workmanager

import (
	"context"
	"fmt"
	"strings"
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

		taskController:     NewTaskController(),
		pipeController:     NewPipeController(ctx),
		poolController:     NewPoolController(ctx),
		limitController:    NewLimitController(ctx),
		callbackController: newCallbackController(),

		mu:             new(sync.RWMutex),
		workerBuilders: make(map[WorkerName]WorkerBuilder, 8),
		stepRunners:    make(map[WorkStep]func(*WorkerManager) error, 8),
	}
}

// WorkerManager worker manager
type WorkerManager struct {
	ctx context.Context

	cacher Cacher

	*taskController
	*pipeController
	*poolController
	*limitController
	*callbackController

	mu             *sync.RWMutex
	workerBuilders map[WorkerName]WorkerBuilder
	stepRunners    map[WorkStep]func(*WorkerManager) error
}

// WithContext set context
func (wm WorkerManager) WithContext(ctx context.Context) *WorkerManager {
	wm.ctx = ctx
	return &wm
}

// SetCacher set cacher
func (wm *WorkerManager) SetCacher(c Cacher) { wm.cacher = c }

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
func (wm *WorkerManager) RegisterStep(current WorkStep, stepRun StepRunner, nexts ...WorkStep) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.stepRunners[current] = func(wm *WorkerManager) error {
		defer wrapPanic("%s step runner panic", current)
		if err := wm.CheckStep(current); err != nil {
			panic(fmt.Errorf("check current step %s fail: %w", current, err))
		}
		for _, step := range nexts {
			if err := wm.CheckStep(step); err != nil {
				panic(fmt.Errorf("check next step %s fail: %w", step, err))
			}
		}

		// get from current's recv channel, stop when step(pipe) removed
		// get limiter every round, in case of limiter updated
		for ch := wm.getRecvChan(current); ch != nil; _ = wm.getLimiter(current).Wait(wm.ctx) {
			select {
			case <-wm.ctx.Done():
				log.CtxInfo(wm.ctx, "step %s runner stopped", current)
				return wm.ctx.Err()
			case target := <-ch:
				wm.run(current, func() {
					defer catchPanic(wm.ctx, "%s step work panic", current)

					task := wm.GetTask(target.Token())
					if wm.cacher != nil && !wm.cacher.Allow(target) {
						return
					}

					callbacks := wm.GetCallbacks(current)

					stepRun(task.Context(),
						wrapWork(task.Context(), callbacks.BeforeWork(), wm.Work, callbacks.AfterWork()),
						target,
						wrapChan(wm.getSendChans(nexts...))...,
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
		for _, call := range after {
			results = call(ctx, results...)
		}
		return results, nil
	}
}

func wrapChan(chs []chan<- WorkTarget) (recvs []func(WorkTarget)) {
	for _, ch := range chs {
		if ch == nil {
			continue
		}

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

// ============ step ============

// InitializeStep initialize step, set goroutine pool, rate limit and pipe for step
func (wm *WorkerManager) InitializeStep(step WorkStep, opts ...PipeOption) {
	if wm.getPool(step) == wm.defaultPool {
		wm.SetPool(0, step)
	}
	if wm.getLimiter(step) == wm.defaultLimiter {
		wm.SetLimit(0, 0, step)
	}
	if !wm.hasPipe(step) {
		wm.InitializePipe(step, opts...)
	}
}

// CheckStep check if step is ready
func (wm *WorkerManager) CheckStep(step WorkStep) error {
	if !wm.hasStep(step) {
		return ErrStepRunnerNotFound
	}

	if err := wm.CheckPipe(step); err != nil {
		return fmt.Errorf("check step's pipe fail: %w", err)
	}

	return nil
}

// RemoveStep remove steps
func (wm *WorkerManager) RemoveStep(steps ...WorkStep) {
	wm.RemovePipe(steps...)
	wm.RemovePool(steps...)

	wm.removeStepRunner(steps...)
}

func (wm *WorkerManager) removeStepRunner(steps ...WorkStep) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	for _, step := range steps {
		delete(wm.stepRunners, step)
	}
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

// HasStep check if has step
func (wm *WorkerManager) hasStep(step WorkStep) bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.stepRunners[step] != nil
}

// StepStatus check step status
func (wm *WorkerManager) StepStatus(step WorkStep) (string, error) {
	// check if step health
	if err := wm.CheckStep(step); err != nil {
		return "", fmt.Errorf("check step %s fail: %w", step, err)
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("step %s's status:\n", step))

	mitm, recvLen, recvCap, sendLen, sendCap := wm.PipeStatus(step)
	buf.WriteString(fmt.Sprintf("pipe status:\n\tmitm: %t\n\trecv chan: [%d/%d]\n\tsend chan: [%d/%d]\n",
		mitm, recvLen, recvCap, sendLen, sendCap))

	num, size := wm.PoolStatus(step)
	buf.WriteString(fmt.Sprintf("pool status: [%d/%d]\n", num, size))

	limit, burst, tokens := wm.LimiterStatus(step)
	buf.WriteString(fmt.Sprintf("limiter status: limit: %.2f, burst:%d, tokens:%.2f\n", limit, burst, tokens))

	before, after := wm.CallbackStatus(step)
	buf.WriteString(fmt.Sprintf("callback status: has %d before hooks, %d after hooks\n", before, after))

	return buf.String(), nil
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

// Serve serve specifid steps, do nothing when steps == nil or len(steps) == 0
func (wm *WorkerManager) Serve(steps ...WorkStep) {
	log.Info("go serve step %v", steps)

	wm.mu.RLock()
	defer wm.mu.RUnlock()
	for _, step := range steps {
		wm.InitializeStep(step)
		// call step run in defer func to avoid panic when step not ready and another step runner call it
		// cause step runner will check step status
		defer func(stepRun func(*WorkerManager) error) { go stepRun(wm) }(wm.stepRunners[step])
	}
}

// Recv put target in specified step's channel
func (wm *WorkerManager) Recv(step WorkStep, target WorkTarget) error {
	if ch := wm.getSendChan(step); ch != nil {
		ch <- target
		return nil
	}
	return fmt.Errorf("%s channel not found", step)
}

// RecvFrom recv from provided channel
// stop waitting data from recv chan when wm.ctx done or recv closed
func (wm *WorkerManager) RecvFrom(step WorkStep, recv <-chan WorkTarget) error {
	if err := wm.CheckStep(step); err != nil {
		return fmt.Errorf("check step %s fail: %w", step, err)
	}
	go func() {
		for {
			select {
			case <-wm.ctx.Done():
				return
			case target, ok := <-recv:
				if !ok { // channel "recv" closed
					log.Debug("new channel for step %s is closed", step)
				} else if err := wm.Recv(step, target); err != nil { // send target
					log.Error("send target to step %s fail: %s", step, err)
				}
			}
		}
	}()
	return nil
}
