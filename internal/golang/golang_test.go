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

package golang

import (
	"sort"
	"strings"
	"testing"

	"github.com/ryanfowler/gopack/internal/types"
)

func envToMap(t *testing.T, env []string) map[string]string {
	t.Helper()
	m := make(map[string]string, len(env))
	for _, e := range env {
		idx := strings.IndexByte(e, '=')
		if idx < 0 {
			t.Fatalf("invalid env entry: %q", e)
		}
		m[e[:idx]] = e[idx+1:]
	}
	return m
}

func TestEnvDeduplicationAndSorting(t *testing.T) {
	t.Setenv("GOOS", "windows")
	t.Setenv("GOARCH", "386")
	t.Setenv("CGO_ENABLED", "1")
	t.Setenv("CUSTOM_VAR", "custom_value")

	b := New()
	platform := types.ParsePlatform("linux/amd64")
	env := b.env(platform)

	if !sort.StringsAreSorted(env) {
		t.Error("env slice is not sorted")
	}

	m := envToMap(t, env)

	if got, want := m["GOOS"], "linux"; got != want {
		t.Errorf("GOOS = %q, want %q", got, want)
	}
	if got, want := m["GOARCH"], "amd64"; got != want {
		t.Errorf("GOARCH = %q, want %q", got, want)
	}
	if got, want := m["CGO_ENABLED"], "0"; got != want {
		t.Errorf("CGO_ENABLED = %q, want %q", got, want)
	}
	if got, want := m["CUSTOM_VAR"], "custom_value"; got != want {
		t.Errorf("CUSTOM_VAR = %q, want %q", got, want)
	}
}

func TestEnvCGOEnabled(t *testing.T) {
	b := New(WithCGOEnabled(true))
	platform := types.ParsePlatform("linux/amd64")
	env := b.env(platform)

	m := envToMap(t, env)
	if got, want := m["CGO_ENABLED"], "1"; got != want {
		t.Errorf("CGO_ENABLED = %q, want %q", got, want)
	}
}

func TestEnvVariants(t *testing.T) {
	tests := []struct {
		platform string
		wantKey  string
		wantVal  string
	}{
		{"linux/arm/v6", "GOARM", "6"},
		{"linux/arm/v7", "GOARM", "7"},
		{"linux/arm:v6", "GOARM", "6"},
		{"linux/arm:v7", "GOARM", "7"},
		{"linux/amd64/v3", "GOAMD64", "v3"},
		{"linux/amd64/v4", "GOAMD64", "v4"},
		{"linux/amd64:v3", "GOAMD64", "v3"},
		{"linux/amd64:v4", "GOAMD64", "v4"},
		{"linux/arm:v8", "", ""},
		{"linux/amd64:v5", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			b := New()
			platform := types.ParsePlatform(tt.platform)
			env := b.env(platform)
			m := envToMap(t, env)

			if tt.wantKey == "" {
				return
			}

			if got, want := m[tt.wantKey], tt.wantVal; got != want {
				t.Errorf("%s = %q, want %q", tt.wantKey, got, want)
			}
		})
	}
}

func TestEnvClearsHostArchitectureTuning(t *testing.T) {
	hostTuning := map[string]string{
		"GO386":     "softfloat",
		"GOAMD64":   "v4",
		"GOARM":     "5,softfloat",
		"GOARM64":   "v9.5,crypto",
		"GOMIPS":    "softfloat",
		"GOMIPS64":  "softfloat",
		"GOPPC64":   "power10",
		"GORISCV64": "rva23u64",
		"GOWASM":    "satconv,signext",
	}
	for key, value := range hostTuning {
		t.Setenv(key, value)
	}

	tests := []struct {
		platform string
		want     map[string]string
	}{
		{
			platform: "linux/386",
			want:     map[string]string{"GO386": "sse2"},
		},
		{
			platform: "linux/amd64",
			want:     map[string]string{"GOAMD64": "v1"},
		},
		{
			platform: "linux/amd64/v3",
			want:     map[string]string{"GOAMD64": "v3"},
		},
		{
			platform: "linux/arm",
			want:     map[string]string{"GOARM": "7"},
		},
		{
			platform: "linux/arm/v5",
			want:     map[string]string{"GOARM": "5"},
		},
		{
			platform: "linux/arm64",
			want:     map[string]string{"GOARM64": "v8.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			b := New()
			platform := types.ParsePlatform(tt.platform)
			env := b.env(platform)
			m := envToMap(t, env)

			for key, want := range tt.want {
				if got := m[key]; got != want {
					t.Errorf("%s = %q, want %q", key, got, want)
				}
			}
			for key, poisoned := range hostTuning {
				if _, ok := tt.want[key]; ok {
					continue
				}
				if got, ok := m[key]; ok {
					t.Errorf("%s leaked as %q from host value %q", key, got, poisoned)
				}
			}
		})
	}
}
