// Copyright 2026 Ryan Fowler
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
	"context"
	"strings"
	"testing"
)

func TestRunRejectsUnsupportedDaemonBeforeOtherWork(t *testing.T) {
	_, err := Run(context.Background(),
		WithDaemon("podman"),
		WithMainPath("/path/that/does/not/exist"),
	)
	if err == nil {
		t.Fatal("Run() error = nil, want unsupported daemon error")
	}
	if !strings.Contains(err.Error(), `unsupported daemon "podman"`) {
		t.Fatalf("Run() error = %q, want unsupported daemon error", err)
	}
}
