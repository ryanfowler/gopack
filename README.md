# gopack

Build and publish a Go binary as a minimal OCI image; Docker not required.

### Install

`gopack` can be installed using the command:

```sh
go install github.com/ryanfowler/gopack@latest
```

Go 1.18+ is required.

### Usage

Use the `gopack run` command to build and publish an image. Although all flags
are optional, some notable flags are:
- `--base`: image to use as the base (default: `gcr.io/distroless/static:nonroot`)
- `--repository`: repository to push the final image to (default: Go binary name)
- `--platform`: platform(s) to build the image(s) for (default: `linux/amd64`)
- `--tag`: tag(s) to push the image with (default: `latest`)
- `--daemon`: push the final image to a local daemon (e.g. `docker`)

As an example: to build an image of the `gopack` CLI for platforms "linux/amd64"
& "linux/arm64" and push it to a Github repository with the tags "1234" &
"5678", you can use the following command from the project root:

```sh
gopack run ./cmd/gopack -r ghcr.io/OWNER/gopack -t 1234 -t 5678 -p linux/amd64 -p linux/arm64
```

Please run `gopack run -h` for more information about the available options.

### License

```
Copyright 2022 Ryan Fowler

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
