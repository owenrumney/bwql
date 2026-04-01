//go:build windows

package table

import (
	"os"
	"strconv"
)

func terminalWidth() int {
	if cols, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && cols > 0 {
		return cols
	}
	return 120
}
