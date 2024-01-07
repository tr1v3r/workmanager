package workmanager

import (
	"errors"
)

var (
	// ErrNilTask task is nil
	ErrNilTask = errors.New("task cannot be nil")

	// ErrStepRunnerNotFound step runner not found
	ErrStepRunnerNotFound = errors.New("step runner not found")
	// ErrPipeNotFound pipe not found
	ErrPipeNotFound = errors.New("pipe not found")

	ErrPipeRecvChanNil = errors.New("pipe's recv chan is nil")
	ErrPipeSendChanNil = errors.New("pipe's send chan is nil")
)
