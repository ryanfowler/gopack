package gopack

import (
	"fmt"
	"os"

	"github.com/ryanfowler/gopack/pkg/types"

	"golang.org/x/term"
)

var (
	_ types.Logger = (*defaultLogger)(nil)
	_ types.Logger = nopLogger{}
)

type defaultLogger struct {
	isTerminal bool
}

func StdErrLogger() types.Logger {
	return &defaultLogger{
		isTerminal: term.IsTerminal(int(os.Stderr.Fd())),
	}
}

func (l *defaultLogger) Println(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
}

func (l *defaultLogger) RePrintf(format string, a ...any) {
	if l.isTerminal {
		fmt.Fprintf(os.Stderr, "\033[2K\r"+format, a...)
	} else {
		l.Println(fmt.Sprintf(format, a...))
	}
}

func (l *defaultLogger) IsNop() bool { return false }

type nopLogger struct{}

func NopLogger() types.Logger { return nopLogger{} }

func (l nopLogger) Println(a ...any) {}

func (l nopLogger) RePrintf(format string, a ...any) {}

func (l nopLogger) IsNop() bool { return true }
