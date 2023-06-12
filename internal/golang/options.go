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
