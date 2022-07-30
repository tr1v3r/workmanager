package workmanager

// StepOption ...
type StepOption func(chan WorkTarget) chan WorkTarget

// StepChSize ...
var StepChSize = func(size int) StepOption {
	return func(ch chan WorkTarget) chan WorkTarget {
		if size < 0 {
			return ch
		}
		return make(chan WorkTarget, size)
	}
}
