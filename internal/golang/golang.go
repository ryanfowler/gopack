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
	"strconv"
	"strings"

	"github.com/ryanfowler/gopack/pkg/types"
)

type GoBuilder struct {
	opts options
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
	var out []string
	out = append(out, os.Environ()...)
	out = append(out, "GOOS="+platform.OS(), "GOARCH="+platform.Arch())
	if b.opts.cgoEnabled {
		out = append(out, "CGO_ENABLED=1")
	} else {
		out = append(out, "CGO_ENABLED=0")
	}

	variantStr := strings.TrimPrefix(platform.Variant(), "v")
	if variantStr == "" {
		return out
	}

	variant, err := strconv.Atoi(variantStr)
	if err != nil {
		return out
	}

	switch platform.Arch() {
	case "arm":
		// https://github.com/golang/go/wiki/MinimumRequirements#arm
		if variant >= 5 && variant <= 7 {
			out = append(out, "GOARM="+variantStr)
		}
	case "amd64":
		// https://github.com/golang/go/wiki/MinimumRequirements#amd64
		if variant >= 1 && variant <= 4 {
			out = append(out, "GOAMD64=v"+variantStr)
		}
	}
	return out
}
