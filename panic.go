package workmanager

import (
	"fmt"
	"runtime"

	"github.com/tr1v3r/pkg/log"
)

// catchPanic catch panic
func catchPanic(format string, args ...interface{}) {
	if e := recover(); e != nil {
		log.Error(format+": %v\n%v", append(args, e, catchStack())...)
	}
}

// catchStack catch stack info
func catchStack() string {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	return fmt.Sprintf("==> %s\n", string(buf[:n]))
}
