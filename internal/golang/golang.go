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
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/ryanfowler/gopack/internal/types"
)

type GoBuilder struct {
	opts options
}

var goArchTuningEnv = []string{
	"GO386",
	"GOAMD64",
	"GOARM",
	"GOARM64",
	"GOMIPS",
	"GOMIPS64",
	"GOPPC64",
	"GORISCV64",
	"GOWASM",
}

func New(options ...Option) *GoBuilder {
	opts := defaultOptions()
	for _, o := range options {
		o(opts)
	}
	return &GoBuilder{opts: *opts}
}

func (b *GoBuilder) GoBuild(ctx context.Context, outPath string, platform types.Platform) error {
	args := []string{"build"}
	if b.opts.trimpathEnabled {
		args = append(args, "-trimpath")
	}
	if b.opts.ldflags != "" {
		args = append(args, "-ldflags", b.opts.ldflags)
	}
	if b.opts.modFlag != "" {
		args = append(args, "-mod", b.opts.modFlag)
	}
	args = append(args, "-o", outPath)
	args = append(args, b.opts.mainPath)

	cmd := exec.CommandContext(ctx, b.opts.goBin, args...)
	cmd.Env = b.env(platform)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := stdout.Bytes()
		if len(msg) == 0 {
			msg = stderr.Bytes()
		}
		return fmt.Errorf("go: %w: %s", err, msg)
	}

	return nil
}

func (b *GoBuilder) env(platform types.Platform) []string {
	envMap := make(map[string]string)
	for _, e := range os.Environ() {
		if before, after, ok := strings.Cut(e, "="); ok {
			envMap[before] = after
		}
	}
	envMap["GOOS"] = platform.OS()
	envMap["GOARCH"] = platform.Arch()
	setTargetArchEnv(envMap, platform)
	envMap["CGO_ENABLED"] = "0"
	if b.opts.cgoEnabled {
		envMap["CGO_ENABLED"] = "1"
	}

	out := make([]string, 0, len(envMap))
	for k, v := range envMap {
		out = append(out, k+"="+v)
	}
	sort.Strings(out)
	return out
}

func setTargetArchEnv(envMap map[string]string, platform types.Platform) {
	for _, key := range goArchTuningEnv {
		delete(envMap, key)
	}

	variant, hasVariant := variantNumber(platform)

	switch platform.Arch() {
	case "386":
		envMap["GO386"] = "sse2"
	case "amd64":
		envMap["GOAMD64"] = "v1"
		if hasVariant && variant >= 1 && variant <= 4 {
			envMap["GOAMD64"] = "v" + strconv.Itoa(variant)
		}
	case "arm":
		envMap["GOARM"] = "7"
		if hasVariant && variant >= 5 && variant <= 7 {
			envMap["GOARM"] = strconv.Itoa(variant)
		}
	case "arm64":
		envMap["GOARM64"] = "v8.0"
	}
}

func variantNumber(platform types.Platform) (int, bool) {
	variantStr := strings.TrimPrefix(platform.Variant(), "v")
	if variantStr != "" {
		if variant, err := strconv.Atoi(variantStr); err == nil {
			return variant, true
		}
	}
	return 0, false
}
