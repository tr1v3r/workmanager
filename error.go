package workmanager

import "errors"

var (
	// ErrNilTask task is nil
	ErrNilTask = errors.New("task cannot be nil")
)
