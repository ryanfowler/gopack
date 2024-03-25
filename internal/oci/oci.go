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

package oci

import (
	"compress/gzip"

	"github.com/ryanfowler/gopack/internal/types"
)

var DefaultTag = "latest"

type BuildOption func(*buildOptions)

func WithCompressionLevel(v int) BuildOption {
	return func(bo *buildOptions) {
		bo.gzipCompressionLevel = v
	}
}

func WithLabels(v map[string]string) BuildOption {
	return func(bo *buildOptions) {
		bo.labels = v
	}
}

type buildOptions struct {
	gzipCompressionLevel int
	labels               map[string]string
}

func defaultBuildOptions() *buildOptions {
	return &buildOptions{
		gzipCompressionLevel: gzip.DefaultCompression,
		labels:               nil,
	}
}

type PushOption func(*pushOptions)

func WithLogger(v types.Logger) PushOption {
	return func(po *pushOptions) {
		po.logger = v
	}
}

func WithTags(v []string) PushOption {
	return func(po *pushOptions) {
		po.tags = v
	}
}

type pushOptions struct {
	logger types.Logger
	tags   []string
}

func defaultPushOptions() *pushOptions {
	return &pushOptions{
		logger: nil,
		tags:   []string{DefaultTag},
	}
}
