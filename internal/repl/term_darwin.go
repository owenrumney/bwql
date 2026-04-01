//go:build darwin

package repl

import (
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

type termState struct {
	Iflag  uint64
	Oflag  uint64
	Cflag  uint64
	Lflag  uint64
	Cc     [20]byte
	Ispeed uint64
	Ospeed uint64
}

func enableRawMode() (*termState, error) {
	var orig termState
	if err := tcget(&orig); err != nil {
		return nil, err
	}

	raw := orig
	raw.Iflag &^= syscall.ICRNL | syscall.IXON
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG | syscall.IEXTEN

	if err := tcset(&raw); err != nil {
		return nil, err
	}

	return &orig, nil
}

func restoreMode(orig *termState) {
	_ = tcset(orig)
}

func isTerminal() bool {
	var t termState
	return tcget(&t) == nil
}

func notifySignals(ch chan<- os.Signal) {
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTSTP)
}

func tcget(t *termState) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		os.Stdin.Fd(),
		uintptr(syscall.TIOCGETA),
		uintptr(unsafe.Pointer(t)), //nolint:gosec // required for terminal ioctl
	)
	if errno != 0 {
		return errno
	}
	return nil
}

func tcset(t *termState) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		os.Stdin.Fd(),
		uintptr(syscall.TIOCSETA),
		uintptr(unsafe.Pointer(t)), //nolint:gosec // required for terminal ioctl
	)
	if errno != 0 {
		return errno
	}
	return nil
}
