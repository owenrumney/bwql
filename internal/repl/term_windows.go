//go:build windows

package repl

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Windows does not support raw terminal mode via syscalls in the same way.
// Fall back to line-buffered input (no arrow key history).

type termState struct{}

func enableRawMode() (*termState, error) {
	return nil, fmt.Errorf("raw mode not supported on Windows")
}

func restoreMode(orig *termState) {}

func isTerminal() bool {
	return false
}

func notifySignals(ch chan<- os.Signal) {
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
}
