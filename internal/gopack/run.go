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

package gopack

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ryanfowler/gopack/internal/golang"
	"github.com/ryanfowler/gopack/internal/oci"
	"github.com/ryanfowler/gopack/internal/types"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	crtypes "github.com/google/go-containerregistry/pkg/v1/types"
	"golang.org/x/sync/errgroup"
)

var ErrNoMatchingImage = errors.New("no matching image")

func Run(ctx context.Context, options ...RunOption) (string, error) {
	opts := defaultRunOptions()
	for _, o := range options {
		o(opts)
	}

	// binName represents the name of the application/binary, as parsed from
	// the provided main path. If no repository is provided, the binName is
	// used.
	binName, err := parseBinName(opts.mainPath)
	if err != nil {
		return "", err
	}
	if opts.repository == "" {
		opts.repository = binName
	}

	platforms := parsePlatforms(opts.logger, opts.platforms)

	baseDesc, err := getBaseDesc(ctx, opts)
	if err != nil {
		return "", err
	}

	baseImgs, err := matchImages(platforms, baseDesc)
	if err != nil {
		return "", err
	}

	imgs, err := buildAllPlatforms(ctx, baseImgs, binName, opts)
	if err != nil {
		return "", err
	}

	output, err := push(ctx, imgs, baseDesc.MediaType, opts)
	if err != nil {
		return "", err
	}

	return output, nil
}

func parseBinName(mainPath string) (string, error) {
	mainPath, err := filepath.Abs(mainPath)
	if err != nil {
		return "", err
	}
	stat, err := os.Stat(mainPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(stat.Name(), filepath.Ext(stat.Name())), nil
}

func buildAllPlatforms(ctx context.Context, imgs map[types.Platform]v1.Image, binName string, opts *runOptions) (map[types.Platform]v1.Image, error) {
	if len(opts.platforms) == 1 {
		opts.logger.Printf("Building image for platform %s\n", opts.platforms[0])
	} else {
		opts.logger.Printf("Building images for platforms %v\n", opts.platforms)
	}

	goBuilder := newGoBuilder(opts)
	semaphore := make(chan struct{}, opts.concurrency)

	var mu sync.Mutex
	out := make(map[types.Platform]v1.Image, len(imgs))

	eg, ctx := errgroup.WithContext(ctx)
	for platform, img := range imgs {
		select {
		case semaphore <- struct{}{}:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		platform := platform
		inImg := img
		eg.Go(func() error {
			defer func() { <-semaphore }()
			outImg, err := build(ctx, goBuilder, binName, platform, inImg, opts)
			if err != nil {
				return fmt.Errorf("building %s: %w", platform, err)
			}
			mu.Lock()
			out[platform] = outImg
			mu.Unlock()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return out, nil
}

func build(ctx context.Context, goBuilder *golang.GoBuilder, binName string, p types.Platform, img v1.Image, opts *runOptions) (v1.Image, error) {
	dir, err := os.MkdirTemp("", "gopack-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	goBinPath := filepath.Join(dir, binName)
	err = goBuilder.GoBuild(ctx, goBinPath, p)
	if err != nil {
		return nil, err
	}

	buildOptions := []oci.BuildOption{
		oci.WithCompressionLevel(opts.compressionLevel),
		oci.WithLabels(opts.labels),
	}
	return oci.BuildImage(ctx, goBinPath, img, buildOptions...)
}

func push(ctx context.Context, imgs map[types.Platform]v1.Image, mt crtypes.MediaType, opts *runOptions) (string, error) {
	if opts.daemon == "docker" {
		if len(imgs) != 1 {
			return "", errors.New("push: can only push a single image to docker")
		}
		var img v1.Image
		for _, i := range imgs {
			img = i
		}
		err := oci.PushDaemon(ctx, opts.repository, img, oci.WithTags(opts.tags), oci.WithLogger(opts.logger))
		if err != nil {
			return "", err
		}
		return chooseOutput(opts.repository, img, opts.tags)
	}

	repo, err := name.NewRepository(opts.repository)
	if err != nil {
		return "", fmt.Errorf("push: parsing repository %q: %w", opts.repository, err)
	}

	if len(imgs) == 1 {
		var img v1.Image
		for _, i := range imgs {
			img = i
		}
		err = oci.Push(ctx, repo, img, oci.WithTags(opts.tags), oci.WithLogger(opts.logger))
		if err != nil {
			return "", err
		}
		return chooseOutput(opts.repository, img, opts.tags)
	}

	addendums := make([]mutate.IndexAddendum, 0, len(imgs))
	for platform, img := range imgs {
		addendums = append(addendums, mutate.IndexAddendum{
			Add: img,
			Descriptor: v1.Descriptor{
				Platform: &v1.Platform{
					Architecture: platform.Arch(),
					OS:           platform.OS(),
					Variant:      platform.Variant(),
				},
			},
		})
	}

	base := mutate.IndexMediaType(empty.Index, mt)
	index := mutate.AppendManifests(base, addendums...)

	err = oci.Push(ctx, repo, index, oci.WithTags(opts.tags))
	if err != nil {
		return "", err
	}
	return chooseOutput(opts.repository, index, opts.tags)
}

type digester interface {
	Digest() (v1.Hash, error)
}

func chooseOutput(repo string, img digester, tags []string) (string, error) {
	var chosen string
	for _, tag := range tags {
		if tag != oci.DefaultTag {
			chosen = tag
			break
		}
	}
	if chosen != "" {
		return fmt.Sprintf("%s:%s", repo, chosen), nil
	}

	digest, err := img.Digest()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s@%s", repo, digest), nil

}

func getBaseDesc(ctx context.Context, opts *runOptions) (*remote.Descriptor, error) {
	opts.logger.Printf("Fetching manifest for base: %s\n", opts.base)
	baseRef, err := name.ParseReference(opts.base)
	if err != nil {
		return nil, fmt.Errorf("unable to parse base: %w", err)
	}
	desc, err := oci.Get(ctx, baseRef)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch base: %w", err)
	}
	return desc, nil
}

func parsePlatforms(l types.Logger, in []string) []types.Platform {
	out := make([]types.Platform, len(in))
	for i, p := range in {
		out[i] = types.ParsePlatform(p)
		if !out[i].IsSupported() {
			l.Printf("Warning: platform %q is not officially supported\n", p)
		}
	}
	return out
}

func matchImages(platforms []types.Platform, desc *remote.Descriptor) (map[types.Platform]v1.Image, error) {
	out := make(map[types.Platform]v1.Image, len(platforms))

	if desc.MediaType.IsImage() {
		for _, platform := range platforms {
			if !platformsEqual(platform, desc.Platform) {
				return nil, fmt.Errorf("base image: platform %s: %w", platform, ErrNoMatchingImage)
			}
			img, err := desc.Image()
			if err != nil {
				return nil, err
			}
			out[platform] = img
		}
		return out, nil
	}

	if desc.MediaType.IsIndex() {
		index, err := desc.ImageIndex()
		if err != nil {
			return nil, err
		}
		manifest, err := index.IndexManifest()
		if err != nil {
			return nil, err
		}
		for _, platform := range platforms {
			desc, ok := matchingDescriptor(platform, manifest.Manifests)
			if !ok {
				return nil, fmt.Errorf("base image: platform %s: %w", platform, ErrNoMatchingImage)
			}
			img, err := index.Image(desc.Digest)
			if err != nil {
				return nil, err
			}
			out[platform] = img
		}
		return out, nil
	}

	return nil, fmt.Errorf("base image: invalid type %q", desc.MediaType)
}

func matchingDescriptor(p types.Platform, ds []v1.Descriptor) (v1.Descriptor, bool) {
	for _, desc := range ds {
		if platformsEqual(p, desc.Platform) {
			return desc, true
		}
	}
	return v1.Descriptor{}, false
}

func platformsEqual(p1 types.Platform, p2 *v1.Platform) bool {
	return p1.OS() == p2.OS && p1.Arch() == p2.Architecture && p1.Variant() == p2.Variant
}

func newGoBuilder(opts *runOptions) *golang.GoBuilder {
	goOptions := []golang.Option{
		golang.WithCGOEnabled(opts.cgoEnabled),
		golang.WithTrimpath(opts.trimpathEnabled),
	}

	if opts.ldflags != "" {
		goOptions = append(goOptions, golang.WithLDFlags(opts.ldflags))
	}
	if opts.mainPath != "" {
		goOptions = append(goOptions, golang.WithMainPath(opts.mainPath))
	}
	if opts.modFlag != "" {
		goOptions = append(goOptions, golang.WithModFlag(opts.modFlag))
	}

	return golang.New(goOptions...)
}
