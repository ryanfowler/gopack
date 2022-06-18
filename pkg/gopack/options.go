package gopack

import (
	"compress/gzip"
	"runtime"

	"github.com/ryanfowler/gopack/internal/oci"
	"github.com/ryanfowler/gopack/pkg/types"
)

type RunOption func(*runOptions)

func WithConcurrency(v int) RunOption {
	return func(ro *runOptions) {
		ro.concurrency = v
	}
}

func WithLogger(v types.Logger) RunOption {
	return func(ro *runOptions) {
		ro.logger = v
	}
}

func WithCGOEnabled(v bool) RunOption {
	return func(ro *runOptions) {
		ro.cgoEnabled = v
	}
}

func WithLDFlags(v string) RunOption {
	return func(ro *runOptions) {
		ro.ldflags = v
	}
}

func WithMainPath(v string) RunOption {
	return func(ro *runOptions) {
		ro.mainPath = v
	}
}

func WithModFlag(v string) RunOption {
	return func(ro *runOptions) {
		ro.modFlag = v
	}
}

func WithTrimpath(v bool) RunOption {
	return func(ro *runOptions) {
		ro.trimpathEnabled = v
	}
}

func WithBase(v string) RunOption {
	return func(ro *runOptions) {
		ro.base = v
	}
}

func WithCompressionLevel(v int) RunOption {
	return func(ro *runOptions) {
		ro.compressionLevel = v
	}
}

func WithDaemon(v string) RunOption {
	return func(ro *runOptions) {
		ro.daemon = v
	}
}

func WithEStargz(v bool) RunOption {
	return func(ro *runOptions) {
		ro.estargzEnabled = v
	}
}

func WithLabels(v map[string]string) RunOption {
	return func(ro *runOptions) {
		ro.labels = v
	}
}

func WithPlatforms(v []string) RunOption {
	return func(ro *runOptions) {
		ro.platforms = v
	}
}

func WithRepository(v string) RunOption {
	return func(ro *runOptions) {
		ro.repository = v
	}
}

func WithTags(v []string) RunOption {
	return func(ro *runOptions) {
		ro.tags = v
	}
}

type runOptions struct {
	// General
	concurrency int
	logger      types.Logger

	// Go
	cgoEnabled      bool
	ldflags         string
	mainPath        string
	modFlag         string
	trimpathEnabled bool

	// Build/Publish
	base             string
	compressionLevel int
	daemon           string
	estargzEnabled   bool
	labels           map[string]string
	platforms        []string
	repository       string
	tags             []string
}

func defaultRunOptions() *runOptions {
	return &runOptions{
		concurrency: runtime.GOMAXPROCS(0),
		logger:      StdErrLogger(),

		cgoEnabled:      false,
		ldflags:         "-s -w",
		mainPath:        ".",
		modFlag:         "",
		trimpathEnabled: true,

		base:             "gcr.io/distroless/static:nonroot",
		compressionLevel: gzip.DefaultCompression,
		daemon:           "",
		estargzEnabled:   false,
		labels:           nil,
		platforms:        []string{types.DefaultPlatform.String()},
		repository:       "",
		tags:             []string{oci.DefaultTag},
	}
}
