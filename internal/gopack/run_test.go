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
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ryanfowler/gopack/internal/types"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
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
