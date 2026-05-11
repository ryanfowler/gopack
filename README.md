# gopack

Build and publish a Go binary as a minimal OCI image

`gopack` uses your locally installed version of Go to cross-compile your
application, building an OCI image with the resulting binary, and pushing your
image to a remote repository or a locally running daemon. Docker not required.

### Install

`gopack` can be installed using the command:

```sh
go install github.com/ryanfowler/gopack/cmd/gopack@latest
```

Go 1.18+ is required.

### Usage

Use `gopack publish` to build and publish an image. `gopack run` is still
supported for compatibility. Although most flags are optional, some notable
flags are:
- `--base`: image to use as the base (default: `gcr.io/distroless/static:nonroot`)
- `--repository`: repository to push the final image to (default: Go binary name)
- `--platform`: platform(s) to build the image(s) for (default: `linux/amd64`)
- `--tag`: tag(s) to push the image with (default: `latest`)
- `--output`: output target for `gopack build` (for example: `oci:./image.tar`)
- `--load`: load the final image to a local daemon
- `--daemon`: local daemon backend (currently: `docker`)

#### Using a custom base image

```sh
gopack publish ./cmd/gopack -b myimage:tag
```

#### Pushing to a specific remote repository

```sh
gopack publish ./cmd/gopack -r ghcr.io/OWNER/gopack
```

#### Building for multiple platforms

```sh
gopack publish ./cmd/gopack -p linux/amd64 -p linux/arm64
```

#### Specifying tags

```sh
gopack publish ./cmd/gopack -t latest -t 12345678
```

#### Build to an OCI archive

```sh
gopack build ./cmd/gopack --output oci:./image.tar
```

#### Push to a local daemon

```sh
gopack load ./cmd/gopack --daemon docker
```

The older daemon form remains valid:

```sh
gopack run ./cmd/gopack --daemon docker
```

_Please run `gopack publish -h`, `gopack build -h`, or `gopack load -h` for more
information about the available options._

### License

```
Copyright 2023 Ryan Fowler

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
