package oci

import (
	"compress/gzip"

	"github.com/ryanfowler/gopack/pkg/types"
)

var DefaultTag = "latest"

type BuildOption func(*buildOptions)

func WithCompressionLevel(v int) BuildOption {
	return func(bo *buildOptions) {
		bo.gzipCompressionLevel = v
	}
}

func WithEStargz(v bool) BuildOption {
	return func(bo *buildOptions) {
		bo.estargzEnabled = v
	}
}

func WithLabels(v map[string]string) BuildOption {
	return func(bo *buildOptions) {
		bo.labels = v
	}
}

type buildOptions struct {
	estargzEnabled       bool
	gzipCompressionLevel int
	labels               map[string]string
}

func defaultBuildOptions() *buildOptions {
	return &buildOptions{
		estargzEnabled:       false,
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
