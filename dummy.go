package workmanager

import "context"

var _ WorkerConfig = new(DummyConfig)

// DummyConfig dummy config
type DummyConfig struct{}

func (c *DummyConfig) Args() map[string]any { return nil }
func (c *DummyConfig) Active() bool         { return true }

var _ Worker = new(DummyWorker)

// DummyWorker dummy worker
type DummyWorker struct{}

func (c *DummyWorker) WithContext(context.Context) Worker       { return c }
func (c *DummyWorker) Work(...WorkTarget) ([]WorkTarget, error) { return nil, nil }

// DummyTarget dummy target
type DummyTarget struct{ TaskToken string }

func (t *DummyTarget) Token() string { return t.TaskToken }
func (t *DummyTarget) Key() string   { return "" }
func (t *DummyTarget) TTL() int      { return 1 }
