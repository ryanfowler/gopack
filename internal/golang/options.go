package golang

type Option func(*options)

func WithCGOEnabled(v bool) Option {
	return func(o *options) {
		o.cgoEnabled = v
	}
}

func WithGoBin(v string) Option {
	return func(o *options) {
		o.goBin = v
	}
}

func WithLDFlags(v string) Option {
	return func(o *options) {
		o.ldflags = v
	}
}

func WithMainPath(v string) Option {
	return func(o *options) {
		o.mainPath = v
	}
}

func WithModFlag(v string) Option {
	return func(o *options) {
		o.modFlag = v
	}
}

func WithTrimpath(v bool) Option {
	return func(o *options) {
		o.trimpathEnabled = v
	}
}

type options struct {
	cgoEnabled      bool
	goBin           string
	ldflags         string
	mainPath        string
	modFlag         string
	trimpathEnabled bool
}

func defaultOptions() *options {
	return &options{
		cgoEnabled:      false,
		goBin:           "go",
		ldflags:         "-s -w",
		mainPath:        ".",
		modFlag:         "",
		trimpathEnabled: true,
	}
}
