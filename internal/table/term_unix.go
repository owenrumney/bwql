//go:build darwin || linux

package table

import (
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

func terminalWidth() int {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}
	var ws winsize
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		os.Stdout.Fd(),
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&ws)), //nolint:gosec // required for terminal ioctl
	)
	if errno != 0 || ws.Col == 0 {
		return terminalWidthFallback()
	}
	return int(ws.Col)
}

func terminalWidthFallback() int {
	if cols, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && cols > 0 {
		return cols
	}
	return 120
}
