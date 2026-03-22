// Package log provides verbose logging for rift CLI operations.
package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

// Logger provides step-by-step verbose output to stderr.
type Logger struct {
	w       io.Writer
	verbose bool
	isTTY   bool
	mu      sync.Mutex
	start   time.Time
}

// New creates a Logger. If verbose is false, all output is suppressed.
func New(verbose bool) *Logger {
	isTTY := term.IsTerminal(int(os.Stderr.Fd()))
	return &Logger{
		w:       os.Stderr,
		verbose: verbose,
		isTTY:   isTTY,
		start:   time.Now(),
	}
}

// Step prints a status message if verbose mode is enabled.
func (l *Logger) Step(msg string) {
	if !l.verbose {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	elapsed := time.Since(l.start).Truncate(time.Millisecond)
	if l.isTTY {
		fmt.Fprintf(l.w, "\033[2m[%s]\033[0m %s\n", elapsed, msg)
	} else {
		fmt.Fprintf(l.w, "[%s] %s\n", elapsed, msg)
	}
}

// Stepf prints a formatted status message if verbose mode is enabled.
func (l *Logger) Stepf(format string, args ...any) {
	l.Step(fmt.Sprintf(format, args...))
}
