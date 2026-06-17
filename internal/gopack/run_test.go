// Copyright 2026 Ryan Fowler
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
	"archive/tar"
	"context"
	"errors"
	"io"
	"log"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ryanfowler/gopack/internal/types"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func TestRunRejectsUnsupportedDaemonBeforeOtherWork(t *testing.T) {
	_, err := Run(context.Background(),
		WithDaemon("podman"),
		WithMainPath("/path/that/does/not/exist"),
	)
	if err == nil {
		t.Fatal("Run() error = nil, want unsupported daemon error")
	}
	if !strings.Contains(err.Error(), `unsupported daemon "podman"`) {
		t.Fatalf("Run() error = %q, want unsupported daemon error", err)
	}
}

func TestRunRejectsUnsupportedPlatformBeforeOtherWork(t *testing.T) {
	_, err := Run(context.Background(),
		WithPlatforms([]string{"linux/ad64"}),
		WithMainPath("/path/that/does/not/exist"),
	)
	if err == nil {
		t.Fatal("Run() error = nil, want unsupported platform error")
	}
	if !strings.Contains(err.Error(), `unsupported platform "linux/ad64"`) {
		t.Fatalf("Run() error = %q, want unsupported platform error", err)
	}
}

func TestRunRejectsUnsupportedOutputBeforeOtherWork(t *testing.T) {
	_, err := Run(context.Background(),
		WithOutput("docker:./image.tar"),
		WithMainPath("/path/that/does/not/exist"),
	)
	if err == nil {
		t.Fatal("Run() error = nil, want unsupported output error")
	}
	if !strings.Contains(err.Error(), `unsupported output "docker:./image.tar"`) {
		t.Fatalf("Run() error = %q, want unsupported output error", err)
	}
}

func TestRunRejectsOutputWithDaemonLoadBeforeOtherWork(t *testing.T) {
	_, err := Run(context.Background(),
		WithOutput("oci:./image.tar"),
		WithDaemon("docker"),
		WithMainPath("/path/that/does/not/exist"),
	)
	if err == nil {
		t.Fatal("Run() error = nil, want conflicting destination error")
	}
	if !strings.Contains(err.Error(), "cannot use output with daemon load") {
		t.Fatalf("Run() error = %q, want conflicting destination error", err)
	}
}

func TestMatchImagesUsesConfigPlatformForImageManifest(t *testing.T) {
	desc := pushImageManifest(t, imageWithPlatform(t, types.ParsePlatform("linux/amd64")))
	if desc.Platform != nil {
		t.Fatalf("test setup: descriptor platform = %#v, want nil", desc.Platform)
	}

	imgs, err := matchImages([]types.Platform{types.ParsePlatform("linux/amd64")}, desc)
	if err != nil {
		t.Fatalf("matchImages() error = %v", err)
	}
	if len(imgs) != 1 {
		t.Fatalf("matchImages() returned %d images, want 1", len(imgs))
	}
}

func TestMatchImagesRejectsImageManifestConfigPlatformMismatch(t *testing.T) {
	desc := pushImageManifest(t, imageWithPlatform(t, types.ParsePlatform("linux/arm64")))

	_, err := matchImages([]types.Platform{types.ParsePlatform("linux/amd64")}, desc)
	if err == nil {
		t.Fatal("matchImages() error = nil, want no matching image error")
	}
	if !errors.Is(err, ErrNoMatchingImage) {
		t.Fatalf("matchImages() error = %v, want %v", err, ErrNoMatchingImage)
	}
}

func TestWriteOCIArchive(t *testing.T) {
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "image.tar")
	imgs := map[types.Platform]v1.Image{
		types.ParsePlatform("linux/amd64"): img,
	}
	if err := writeOCIArchive(path, imgs); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	seen := map[string]bool{}
	tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		seen[hdr.Name] = true
	}

	for _, name := range []string{"oci-layout", "index.json"} {
		if !seen[name] {
			t.Fatalf("OCI archive missing %s", name)
		}
	}
}

func imageWithPlatform(t *testing.T, platform types.Platform) v1.Image {
	t.Helper()

	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatal(err)
	}
	config, err := img.ConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	config = config.DeepCopy()
	config.OS = platform.OS()
	config.Architecture = platform.Arch()
	config.Variant = strings.TrimPrefix(platform.Variant(), "v")

	img, err = mutate.ConfigFile(img, config)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func pushImageManifest(t *testing.T, img v1.Image) *remote.Descriptor {
	t.Helper()

	server := httptest.NewServer(registry.New(registry.Logger(log.New(io.Discard, "", 0))))
	t.Cleanup(server.Close)

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := name.NewTag(u.Host+"/base:latest", name.Insecure)
	if err != nil {
		t.Fatal(err)
	}
	if err := remote.Write(ref, img); err != nil {
		t.Fatal(err)
	}
	desc, err := remote.Get(ref)
	if err != nil {
		t.Fatal(err)
	}
	return desc
}
