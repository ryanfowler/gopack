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
