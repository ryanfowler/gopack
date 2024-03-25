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
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

func BuildImage(ctx context.Context, goBinPath string, base v1.Image, options ...BuildOption) (v1.Image, error) {
	opts := defaultBuildOptions()
	for _, o := range options {
		o(opts)
	}

	entrypoint := "/app/" + filepath.Base(goBinPath)
	raw, err := tarGoBin(goBinPath, entrypoint)
	if err != nil {
		return nil, fmt.Errorf("tar go binary: %w", err)
	}

	layerOpts := []tarball.LayerOption{
		tarball.WithCompressedCaching,
		tarball.WithCompressionLevel(opts.gzipCompressionLevel),
	}

	layer, err := tarball.LayerFromOpener(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(raw)), nil
	}, layerOpts...)
	if err != nil {
		return nil, err
	}

	out, err := mutate.Append(base, mutate.Addendum{
		Layer: layer,
		History: v1.History{
			Author:    "gopack",
			Created:   v1.Time{},
			CreatedBy: "gopack run ...",
		},
	})
	if err != nil {
		return nil, err
	}

	config, err := out.ConfigFile()
	if err != nil {
		return nil, err
	}
	config = config.DeepCopy()

	config.Author = "gopack"
	config.Config.Cmd = nil
	config.Config.Entrypoint = []string{entrypoint}
	if config.Config.Labels == nil && len(opts.labels) > 0 {
		config.Config.Labels = make(map[string]string)
	}
	for key, val := range opts.labels {
		config.Config.Labels[key] = val
	}

	out, err = mutate.ConfigFile(out, config)
	if err != nil {
		return nil, err
	}

	return mutate.MediaType(out, types.DockerManifestSchema2), nil
}

func tarGoBin(goBinPath, entrypoint string) ([]byte, error) {
	file, err := os.Open(goBinPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	err = tw.WriteHeader(&tar.Header{
		Mode:     0o555,
		Name:     strings.TrimPrefix(entrypoint, "/"),
		Size:     stat.Size(),
		Typeflag: tar.TypeReg,
	})
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(tw, file); err != nil {
		return nil, err
	}
	if err = tw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
