package workmanager

import "context"

var _ WorkConfig = new(DummyConfig)

// DummyConfig dummy config
type DummyConfig struct{}

func (c *DummyConfig) Args() map[string]interface{} { return nil }
func (c *DummyConfig) Active() bool                 { return true }

var _ Worker = new(DummyWorker)

// DummyWorker dummy worker
type DummyWorker struct{}

func (c *DummyWorker) LoadConfig(WorkerConfig) Worker      { return c }
func (c *DummyWorker) WithContext(context.Context) Worker  { return c }
func (c *DummyWorker) GetContext() context.Context         { return nil }
func (c *DummyWorker) BeforeWork()                         {}
func (c *DummyWorker) Work(WorkTarget) (WorkTarget, error) { return nil, nil }
func (c *DummyWorker) AfterWork()                          {}
func (c *DummyWorker) GetResult() WorkTarget               { return nil }
func (c *DummyWorker) Finished() <-chan struct{}           { return nil }
func (c *DummyWorker) Terminate() error                    { return nil }
