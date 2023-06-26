package workmanager

import "context"

var (
	// TransferRunner runner for transfering
	TransferRunner = func(trans func(WorkTarget)) StepRunner {
		return func(_ context.Context, _ Work, target WorkTarget, _ ...func(WorkTarget)) { trans(target) }
	}
)
