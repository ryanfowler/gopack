// Copyright 2023 Ryan Fowler
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// StdErrLogger returns a new Logger that logs to stderr.
func StdErrLogger() types.Logger {
	return &defaultLogger{
		isTerminal: term.IsTerminal(int(os.Stderr.Fd())),
	}
}

func (l *defaultLogger) Printf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func (l *defaultLogger) Println(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
}

func (l *defaultLogger) RePrintf(format string, a ...any) {
	if l.isTerminal {
		fmt.Fprintf(os.Stderr, "\033[2K\r"+format, a...)
	}
}

type nopLogger struct{}

// NopLogger returns a Logger with no-op methods.
func NopLogger() types.Logger { return nopLogger{} }

func (l nopLogger) Printf(format string, a ...any) {}

func (l nopLogger) Println(a ...any) {}

func (l nopLogger) RePrintf(format string, a ...any) {}
