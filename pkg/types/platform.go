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
