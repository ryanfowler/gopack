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

import "strings"

var DefaultPlatform = Platform{os: "linux", arch: "amd64"}

type Platform struct {
	os      string
	arch    string
	variant string
}

func ParsePlatform(s string) Platform {
	var out Platform

	if idx := strings.Index(s, ":"); idx >= 0 {
		out.variant = s[idx+1:]
		s = s[:idx]
	}
	if idx := strings.Index(s, "/"); idx >= 0 {
		out.arch = s[idx+1:]
		s = s[:idx]
	}
	out.os = s

	return out
}

func (p Platform) OS() string {
	return p.os
}

func (p Platform) Arch() string {
	return p.arch
}

func (p Platform) Variant() string {
	return p.variant
}

func (p Platform) IsEqual(pf Platform) bool {
	return p.os == pf.os && p.arch == pf.arch && p.variant == pf.variant
}

func (p Platform) String() string {
	out := p.os + "/" + p.arch
	if p.variant != "" {
		out += ":" + p.variant
	}
	return out
}

func (p Platform) IsSupported() bool {
	for _, platform := range supportedPlatforms {
		if p.os != platform.os || p.arch != platform.arch {
			continue
		}
		for _, variant := range platform.variants {
			if strings.TrimPrefix(p.variant, "v") == variant {
				return true
			}
		}
	}
	return false
}

type platform struct {
	os       string
	arch     string
	variants []string
}

var supportedPlatforms = []platform{
	{os: "linux", arch: "amd64", variants: []string{"", "1", "2", "3", "4"}},
	{os: "linux", arch: "arm64", variants: []string{""}},
	{os: "linux", arch: "386", variants: []string{""}},
	{os: "linux", arch: "arm", variants: []string{"", "5", "6", "7"}},
}
