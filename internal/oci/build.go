package oci

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/containerd/stargz-snapshotter/estargz"
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
	if opts.estargzEnabled {
		layerOpts = append(layerOpts,
			tarball.WithEstargz,
			tarball.WithEstargzOptions(
				estargz.WithPrioritizedFiles([]string{entrypoint})))
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
		Mode:     0555,
		Name:     entrypoint,
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
