//go:build linux

package repl

import (
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

type termState struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Line   uint8
	Cc     [32]byte
	Ispeed uint32
	Ospeed uint32
}

const (
	tcgets = 0x5401
	tcsets = 0x5402
)

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
		uintptr(syscall.Stdin),  //nolint:gosec // int->uintptr required for syscall
		uintptr(tcgets),
		uintptr(unsafe.Pointer(t)), //nolint:gosec // unsafe pointer required for ioctl
	)
	if errno != 0 {
		return errno
	}
	return nil
}

func tcset(t *termState) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),  //nolint:gosec // int->uintptr required for syscall
		uintptr(tcsets),
		uintptr(unsafe.Pointer(t)), //nolint:gosec // unsafe pointer required for ioctl
	)
	if errno != 0 {
		return errno
	}
	return nil
}
