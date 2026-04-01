package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"
)

type Readline struct {
	prompt  string
	history []string
	histIdx int
	scanner *bufio.Scanner
}

func New(prompt string) *Readline {
	return &Readline{prompt: prompt}
}

func (r *Readline) ReadLine() (string, error) {
	if !isTerminal() {
		return r.readLineFallback()
	}

	orig, err := enableRawMode()
	if err != nil {
		return r.readLineFallback()
	}

	// Ensure terminal is restored on signals
	sigCh := make(chan os.Signal, 1)
	notifySignals(sigCh)
	defer signal.Stop(sigCh)
	go func() {
		sig, ok := <-sigCh
		if ok {
			restoreMode(orig)
			// Re-raise so the default handler runs
			signal.Reset(sig)
			p, _ := os.FindProcess(os.Getpid())
			_ = p.Signal(sig)
		}
	}()

	defer restoreMode(orig)

	r.histIdx = len(r.history)

	var buf []byte
	pos := 0

	r.writePrompt()

	b := make([]byte, 1)
	for {
		_, err := os.Stdin.Read(b)
		if err != nil {
			if err == io.EOF {
				return "", io.EOF
			}
			return "", err
		}

		switch b[0] {
		case 3: // Ctrl-C
			writeOut("^C\r\n")
			return "", io.EOF

		case 4: // Ctrl-D
			if len(buf) == 0 {
				writeOut("\r\n")
				return "", io.EOF
			}

		case 13, 10: // Enter
			writeOut("\r\n")
			line := string(buf)
			if line != "" {
				r.history = append(r.history, line)
				if len(r.history) > 1000 {
					r.history = r.history[len(r.history)-500:]
				}
			}
			return line, nil

		case 127, 8: // Backspace
			if pos > 0 {
				buf = append(buf[:pos-1], buf[pos:]...)
				pos--
				r.refreshLine(buf, pos)
			}

		case 27: // Escape sequence
			seq, ok := r.readEscSeq(2)
			if !ok {
				continue // bare Escape — ignore
			}
			if seq[0] == '[' {
				switch seq[1] {
				case 'A': // Up
					if r.histIdx > 0 {
						r.histIdx--
						buf = []byte(r.history[r.histIdx])
						pos = len(buf)
						r.refreshLine(buf, pos)
					}
				case 'B': // Down
					if r.histIdx < len(r.history)-1 {
						r.histIdx++
						buf = []byte(r.history[r.histIdx])
						pos = len(buf)
						r.refreshLine(buf, pos)
					} else if r.histIdx == len(r.history)-1 {
						r.histIdx = len(r.history)
						buf = nil
						pos = 0
						r.refreshLine(buf, pos)
					}
				case 'C': // Right
					if pos < len(buf) {
						pos++
						r.refreshLine(buf, pos)
					}
				case 'D': // Left
					if pos > 0 {
						pos--
						r.refreshLine(buf, pos)
					}
				case '3': // Delete key (escape [ 3 ~)
					extra, ok := r.readEscSeq(1)
					if ok && extra[0] == '~' && pos < len(buf) {
						buf = append(buf[:pos], buf[pos+1:]...)
						r.refreshLine(buf, pos)
					}
				}
			}

		case 1: // Ctrl-A (home)
			pos = 0
			r.refreshLine(buf, pos)

		case 5: // Ctrl-E (end)
			pos = len(buf)
			r.refreshLine(buf, pos)

		case 21: // Ctrl-U (clear line)
			buf = nil
			pos = 0
			r.refreshLine(buf, pos)

		case 12: // Ctrl-L (clear screen)
			writeOut("\x1b[2J\x1b[H")
			r.refreshLine(buf, pos)

		default:
			if b[0] >= 32 {
				if pos == len(buf) {
					buf = append(buf, b[0])
				} else {
					buf = append(buf[:pos+1], buf[pos:]...)
					buf[pos] = b[0]
				}
				pos++
				r.refreshLine(buf, pos)
			}
		}
	}
}

// readEscSeq reads n bytes with a short timeout.
// Returns the bytes and true if all were read, or false on timeout.
func (r *Readline) readEscSeq(n int) ([]byte, bool) {
	buf := make([]byte, n)
	done := make(chan int, 1)
	go func() {
		read, _ := io.ReadFull(os.Stdin, buf)
		done <- read
	}()
	select {
	case count := <-done:
		return buf, count == n
	case <-time.After(50 * time.Millisecond):
		return nil, false
	}
}

func (r *Readline) writePrompt() {
	writeOut(r.prompt)
}

func (r *Readline) refreshLine(buf []byte, pos int) {
	writeOut("\r\x1b[K")
	writeOut(r.prompt)
	_, _ = os.Stdout.Write(buf)
	cursorPos := len(r.prompt) + pos
	_, _ = fmt.Fprintf(os.Stdout, "\r\x1b[%dC", cursorPos)
}

func writeOut(s string) {
	_, _ = os.Stdout.WriteString(s)
}

func (r *Readline) readLineFallback() (string, error) {
	if r.scanner == nil {
		r.scanner = bufio.NewScanner(os.Stdin)
	}
	fmt.Print(r.prompt)
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return r.scanner.Text(), nil
}
