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

package types

import (
	"testing"
)

func TestPlatform(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expOut      Platform
		isSupported bool
	}{
		{
			name:        "should parse default platform",
			input:       DefaultPlatform.String(),
			expOut:      DefaultPlatform,
			isSupported: true,
		},
		{
			name:        "should parse platform with variant",
			input:       "linux/amd64:v4",
			expOut:      Platform{os: "linux", arch: "amd64", variant: "v4"},
			isSupported: true,
		},
		{
			name:        "should be unsupported arch",
			input:       "linux/bad",
			expOut:      Platform{os: "linux", arch: "bad"},
			isSupported: false,
		},
		{
			name:        "should be unsupported os",
			input:       "windows/arm64",
			expOut:      Platform{os: "windows", arch: "arm64"},
			isSupported: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			p := ParsePlatform(test.input)
			if !p.IsEqual(test.expOut) {
				t.Fatalf("Unexpected platform: %+v", p)
			}
			if ok := p.IsSupported(); ok != test.isSupported {
				t.Fatalf("Unexpected IsSupported result: %v", ok)
			}
		})
	}
}
