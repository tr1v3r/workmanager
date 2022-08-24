package workmanager

var (
	// TransferRunner runner for transfering
	TransferRunner = func(trans func(WorkTarget)) StepRunner {
		return func(_ Work, target WorkTarget, _ ...func(WorkTarget)) { trans(target) }
	}
)
