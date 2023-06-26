package workmanager

import "context"

var (
	// TransferRunner runner for transfering
	TransferRunner = func(trans func(context.Context, WorkTarget)) StepRunner {
		return func(ctx context.Context, _ Work, target WorkTarget, _ ...func(WorkTarget)) { trans(ctx, target) }
	}
)
