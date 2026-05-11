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
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/ryanfowler/gopack/internal/gopack"
	"github.com/ryanfowler/gopack/internal/oci"
	"github.com/ryanfowler/gopack/internal/types"

	"github.com/spf13/cobra"
)

const dockerDaemon = "docker"

type commandMode int

const (
	modeRun commandMode = iota
	modePublish
	modeBuild
	modeLoad
)

type cliOptions struct {
	base        string
	cgoEnabled  bool
	compression int
	concurrency int
	daemon      string
	labels      []string
	ldflags     string
	load        bool
	mod         string
	output      string
	platforms   []string
	repository  string
	tags        []string
	trimpath    bool
}

var rootCmd = newRootCmd()

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "gopack",
		SilenceUsage: true,
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		cmd.Version = info.Main.Version
	}

	cmd.AddCommand(
		newRunCommand(),
		newPublishCommand(),
		newBuildCommand(),
		newLoadCommand(),
	)
	return cmd
}

func newRunCommand() *cobra.Command {
	opts := defaultCLIOptions()
	cmd := newPackageCommand(modeRun, opts)
	cmd.Use = "run [package]"
	cmd.Short = "Build and publish or load a Go binary as a minimal OCI image"
	addCommonFlags(cmd, opts)
	addLoadFlags(cmd, opts)
	addDaemonFlag(cmd, opts, "")
	return cmd
}

func newPublishCommand() *cobra.Command {
	opts := defaultCLIOptions()
	cmd := newPackageCommand(modePublish, opts)
	cmd.Use = "publish [package]"
	cmd.Short = "Build and publish a Go binary as a minimal OCI image"
	addCommonFlags(cmd, opts)
	return cmd
}

func newBuildCommand() *cobra.Command {
	opts := defaultCLIOptions()
	cmd := newPackageCommand(modeBuild, opts)
	cmd.Use = "build [package]"
	cmd.Short = "Build a Go binary as a minimal OCI image archive"
	addCommonFlags(cmd, opts)
	cmd.Flags().StringVarP(&opts.output, "output", "o", "", "output target (e.g. oci:./image.tar)")
	return cmd
}

func newLoadCommand() *cobra.Command {
	opts := defaultCLIOptions()
	cmd := newPackageCommand(modeLoad, opts)
	cmd.Use = "load [package]"
	cmd.Short = "Build and load a Go binary image into a local daemon"
	addCommonFlags(cmd, opts)
	addDaemonFlag(cmd, opts, dockerDaemon)
	return cmd
}

func newPackageCommand(mode commandMode, opts *cliOptions) *cobra.Command {
	return &cobra.Command{
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateCommandOptions(mode, opts); err != nil {
				return err
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			options, err := buildRunOptions(mode, opts, args)
			if err != nil {
				return err
			}

			out, err := gopack.Run(ctx, options...)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), out)
			return nil
		},
	}
}

func defaultCLIOptions() *cliOptions {
	return &cliOptions{
		base:        "gcr.io/distroless/static:nonroot",
		compression: -1,
		platforms:   []string{types.DefaultPlatform.String()},
		tags:        []string{oci.DefaultTag},
		trimpath:    true,
	}
}

func addCommonFlags(cmd *cobra.Command, opts *cliOptions) {
	cmd.Flags().StringVarP(&opts.base, "base", "b", opts.base, "repository to use as the base image")
	cmd.Flags().BoolVar(&opts.cgoEnabled, "cgo", opts.cgoEnabled, "enable CGO during Go compilation")
	cmd.Flags().IntVar(&opts.compression, "compression", opts.compression, "gzip compression level of image layers")
	cmd.Flags().IntVarP(&opts.concurrency, "concurrency", "c", opts.concurrency, "number of concurrent builds (default GOMAXPROCS)")
	cmd.Flags().StringSliceVarP(&opts.labels, "label", "l", opts.labels, "labels to include in image")
	cmd.Flags().StringVar(&opts.ldflags, "ldflags", opts.ldflags, "ldflags used during Go compilation")
	cmd.Flags().StringVar(&opts.mod, "mod", opts.mod, "mod flag used during Go compilation")
	cmd.Flags().StringSliceVarP(&opts.platforms, "platform", "p", opts.platforms, "platforms to build for")
	cmd.Flags().StringVarP(&opts.repository, "repository", "r", opts.repository, "repository to name or push image as")
	cmd.Flags().StringSliceVarP(&opts.tags, "tag", "t", opts.tags, "tags to apply to the image")
	cmd.Flags().BoolVar(&opts.trimpath, "trimpath", opts.trimpath, "enable trimpath during Go compilation")
}

func addLoadFlags(cmd *cobra.Command, opts *cliOptions) {
	cmd.Flags().BoolVar(&opts.load, "load", false, "load image to a local daemon")
}

func addDaemonFlag(cmd *cobra.Command, opts *cliOptions, daemonDefault string) {
	opts.daemon = daemonDefault
	cmd.Flags().StringVarP(&opts.daemon, "daemon", "d", daemonDefault, "local daemon backend (supported: docker)")
}

func validateCommandOptions(mode commandMode, opts *cliOptions) error {
	switch mode {
	case modeBuild:
		if opts.output == "" {
			return fmt.Errorf("build requires --output")
		}
	case modeLoad:
		opts.load = true
	case modeRun:
		if opts.output != "" {
			return fmt.Errorf("run does not support --output; use build")
		}
	}
	return nil
}

func buildRunOptions(mode commandMode, opts *cliOptions, args []string) ([]gopack.RunOption, error) {
	options := []gopack.RunOption{
		gopack.WithCGOEnabled(opts.cgoEnabled),
		gopack.WithTrimpath(opts.trimpath),
	}

	if opts.concurrency > 0 {
		options = append(options, gopack.WithConcurrency(opts.concurrency))
	}
	if opts.ldflags != "" {
		options = append(options, gopack.WithLDFlags(opts.ldflags))
	}
	if len(args) == 1 {
		options = append(options, gopack.WithMainPath(args[0]))
	}
	if opts.mod != "" {
		options = append(options, gopack.WithModFlag(opts.mod))
	}
	if opts.base != "" {
		options = append(options, gopack.WithBase(opts.base))
	}
	if opts.compression >= 0 {
		options = append(options, gopack.WithCompressionLevel(opts.compression))
	}
	if opts.daemon != "" {
		options = append(options, gopack.WithDaemon(opts.daemon))
	}
	if opts.load {
		options = append(options, gopack.WithLoad(true))
	}
	if opts.output != "" {
		options = append(options, gopack.WithOutput(opts.output))
	}
	if len(opts.labels) > 0 {
		m, err := parseLabels(opts.labels)
		if err != nil {
			return nil, err
		}
		options = append(options, gopack.WithLabels(m))
	}
	if len(opts.platforms) > 0 {
		options = append(options, gopack.WithPlatforms(opts.platforms))
	}
	if opts.repository != "" {
		options = append(options, gopack.WithRepository(opts.repository))
	}
	if len(opts.tags) > 0 {
		options = append(options, gopack.WithTags(opts.tags))
	}
	if mode == modeLoad {
		options = append(options, gopack.WithLoad(true))
	}

	return options, nil
}

func parseLabels(labels []string) (map[string]string, error) {
	m := make(map[string]string, len(labels))
	for _, label := range labels {
		var key, val string
		idx := strings.Index(label, "=")
		if idx >= 0 {
			key = label[:idx]
			val = label[idx+1:]
		} else {
			key = label
		}
		if key == "" {
			return nil, fmt.Errorf("invalid label %q: empty key", label)
		}
		if !isValidLabelKey(key) {
			return nil, fmt.Errorf("invalid label key %q", key)
		}
		m[key] = val
	}
	return m, nil
}

func isValidLabelKey(key string) bool {
	if key == "" {
		return false
	}
	for i, c := range key {
		if i == 0 && !isAlphaNum(c) {
			return false
		}
		if !isAlphaNum(c) && c != '.' && c != '_' && c != '-' && c != '/' {
			return false
		}
	}
	return true
}

func isAlphaNum(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
