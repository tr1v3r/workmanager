package workmanager

// DummyConfig dummy config
type DummyConfig struct{}

func (c *DummyConfig) Args() map[string]interface{} { return nil }
func (c *DummyConfig) Active() bool                 { return true }
