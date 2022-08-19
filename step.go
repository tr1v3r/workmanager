package workmanager

// PipeOption ...
type PipeOption func(chan WorkTarget) chan WorkTarget

// StepChSize ...
var StepChSize = func(size int) PipeOption {
	return func(ch chan WorkTarget) chan WorkTarget {
		if size < 0 {
			return ch
		}
		return make(chan WorkTarget, size)
	}
}

// StepRecver ...
var StepRecver = func(recv func(WorkTarget)) PipeOption {
	return func(ch chan WorkTarget) chan WorkTarget {
		go func() {
			select {
			case target := <-ch:
				recv(target)
			}
		}()
		return ch
	}
}
