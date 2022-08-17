package workmanager

func ContainsStep(step WorkStep, steps ...WorkStep) bool {
	if len(steps) == 0 {
		return false
	}

	for _, s := range steps {
		if s == step {
			return true
		}
	}
	return false
}
