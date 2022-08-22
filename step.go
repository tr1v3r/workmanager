package workmanager

// PipeOption ...
type PipeOption func(chan WorkTarget) chan WorkTarget

// PipeChSize ...
var PipeChSize = func(size int) PipeOption {
	return func(ch chan WorkTarget) chan WorkTarget {
		if size < 0 {
			return ch
		}
		return make(chan WorkTarget, size)
	}
}

var (
	// TransRunner runner for transfering
	TransRunner = func(trans func(WorkTarget)) StepRunner {
		return func(_ Work, target WorkTarget, _ ...func(WorkTarget)) { trans(target) }
	}
)
