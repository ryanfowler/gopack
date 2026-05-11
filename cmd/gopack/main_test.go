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

package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestBuildRequiresOutput(t *testing.T) {
	_, err := executeCommand("build", "/path/that/does/not/exist")
	if err == nil {
		t.Fatal("command error = nil, want missing output error")
	}
	if !strings.Contains(err.Error(), "build requires --output") {
		t.Fatalf("command error = %q, want missing output error", err)
	}
}

func TestLoadRejectsUnsupportedDaemonBeforeOtherWork(t *testing.T) {
	_, err := executeCommand("load", "/path/that/does/not/exist", "--daemon", "podman")
	if err == nil {
		t.Fatal("command error = nil, want unsupported daemon error")
	}
	if !strings.Contains(err.Error(), `unsupported daemon "podman"`) {
		t.Fatalf("command error = %q, want unsupported daemon error", err)
	}
}

func executeCommand(args ...string) (string, error) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestIsValidLabelKey(t *testing.T) {
	tests := []struct {
		key   string
		valid bool
	}{
		{"", false},
		{"a", true},
		{"A", true},
		{"0", true},
		{"key", true},
		{"key-name", true},
		{"key_name", true},
		{"key.name", true},
		{"key/name", true},
		{"org.opencontainers.image.title", true},
		{"com.example.myKey", true},
		{"-key", false},
		{"_key", false},
		{".key", false},
		{"/key", false},
		{"key space", false},
		{"key\ttab", false},
		{"key\nnewline", false},
		{"key!", false},
		{"key@", false},
		{"key#", false},
		{"key$", false},
		{"key%", false},
		{"key^", false},
		{"key&", false},
		{"key*", false},
		{"key(", false},
		{"key)", false},
		{"key+", false},
		{"key=", false},
		{"key[", false},
		{"key]", false},
		{"key{", false},
		{"key}", false},
		{"key|", false},
		{"key\\", false},
		{"key:", false},
		{"key;", false},
		{"key<", false},
		{"key>", false},
		{"key?", false},
		{"key'", false},
		{"key\"", false},
		{"key~", false},
		{"key`", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := isValidLabelKey(tt.key)
			if got != tt.valid {
				t.Errorf("isValidLabelKey(%q) = %v, want %v", tt.key, got, tt.valid)
			}
		})
	}
}
