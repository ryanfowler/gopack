// Copyright 2022 Ryan Fowler
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
	"runtime/debug"
	"strings"

	"github.com/ryanfowler/gopack/internal/oci"
	"github.com/ryanfowler/gopack/pkg/gopack"
	"github.com/ryanfowler/gopack/pkg/types"

	"github.com/spf13/cobra"
)

var (
	base        string
	cgoEnabled  bool
	compression int
	concurrency int
	daemon      string
	estargz     bool
	labels      []string
	ldflags     string
	mod         string
	platforms   []string
	repository  string
	tags        []string
	trimpath    bool
)

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		rootCmd.Version = info.Main.Version
	}

	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&base, "base", "b", "gcr.io/distroless/static:nonroot", "repository to use as the base image")
	runCmd.Flags().BoolVar(&cgoEnabled, "cgo", false, "enable CGO during Go compilation")
	runCmd.Flags().IntVar(&compression, "compression", -1, "gzip compression level of image layers")
	runCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 0, "number of concurrent builds (default GOMAXPROCS)")
	runCmd.Flags().StringVarP(&daemon, "daemon", "d", "", "push image to local daemon (e.g. docker)")
	runCmd.Flags().BoolVar(&estargz, "estargz", false, "enable estargz on image")
	runCmd.Flags().StringSliceVarP(&labels, "label", "l", nil, "labels to include in image")
	runCmd.Flags().StringVar(&ldflags, "ldflags", "", "ldflags used during Go compilation")
	runCmd.Flags().StringVar(&mod, "mod", "", "mod flag used during Go compilation")
	runCmd.Flags().StringSliceVarP(&platforms, "platform", "p", []string{types.DefaultPlatform.String()}, "platforms to build for")
	runCmd.Flags().StringVarP(&repository, "repository", "r", "", "repository to push image to")
	runCmd.Flags().StringSliceVarP(&tags, "tag", "t", []string{oci.DefaultTag}, "tags to push image with")
	runCmd.Flags().BoolVar(&trimpath, "trimpath", true, "enable trimpath during Go compilation")
}

var rootCmd = &cobra.Command{
	Use:          "gopack",
	SilenceUsage: true,
}

var runCmd = &cobra.Command{
	Use:   "run [package]",
	Args:  cobra.RangeArgs(0, 1),
	Short: "Build and publish a Go binary as a minimal OCI image",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		options := []gopack.RunOption{
			gopack.WithCGOEnabled(cgoEnabled),
			gopack.WithTrimpath(trimpath),
			gopack.WithEStargz(estargz),
		}

		if concurrency > 0 {
			options = append(options, gopack.WithConcurrency(concurrency))
		}
		if ldflags != "" {
			options = append(options, gopack.WithLDFlags(ldflags))
		}
		if len(args) == 1 {
			options = append(options, gopack.WithMainPath(args[0]))
		}
		if mod != "" {
			options = append(options, gopack.WithModFlag(mod))
		}
		if base != "" {
			options = append(options, gopack.WithBase(base))
		}
		if compression >= 0 {
			options = append(options, gopack.WithCompressionLevel(compression))
		}
		if daemon != "" {
			options = append(options, gopack.WithDaemon(daemon))
		}
		if len(labels) > 0 {
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
				m[key] = val
			}
			options = append(options, gopack.WithLabels(m))
		}
		if len(platforms) > 0 {
			options = append(options, gopack.WithPlatforms(platforms))
		}
		if repository != "" {
			options = append(options, gopack.WithRepository(repository))
		}
		if len(tags) > 0 {
			options = append(options, gopack.WithTags(tags))
		}

		out, err := gopack.Run(ctx, options...)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
