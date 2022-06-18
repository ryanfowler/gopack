// Copyright 2022 Ryan Fowler
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
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/github"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/ryanfowler/gopack/pkg/types"
)

var keychain = authn.NewMultiKeychain(
	authn.DefaultKeychain,
	github.Keychain,
)

func Get(ctx context.Context, ref name.Reference) (*remote.Descriptor, error) {
	return remote.Get(ref,
		remote.WithContext(ctx),
		remote.WithAuthFromKeychain(keychain))
}

func PushDaemon(ctx context.Context, imgName string, img v1.Image, options ...PushOption) error {
	opts := defaultPushOptions()
	for _, o := range options {
		o(opts)
	}

	digest, err := img.Digest()
	if err != nil {
		return err
	}
	srcTag, err := name.NewTag(imgName + ":" + digest.Hex)
	if err != nil {
		return err
	}

	opts.logger.Println(fmt.Sprintf("Pushing digest %s", digest.Hex))
	_, err = daemon.Write(srcTag, img, daemon.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("pushing to daemon: %s: %w", srcTag.String(), err)
	}

	for _, raw := range opts.tags {
		tag, err := name.NewTag(imgName + ":" + raw)
		if err != nil {
			return fmt.Errorf("pushing to daemon: invalid tag: %s: %w", tag.String(), err)
		}
		opts.logger.Println(fmt.Sprintf("Pushing tag %s", raw))
		err = daemon.Tag(srcTag, tag, daemon.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("pushing to daemon: %s: %w", tag.String(), err)
		}
	}

	return nil
}

func Push(ctx context.Context, repo name.Repository, img remote.Taggable, options ...PushOption) error {
	opts := defaultPushOptions()
	for _, o := range options {
		o(opts)
	}

	if len(opts.tags) == 0 {
		return errors.New("push: no tags provided")
	}

	for i, raw := range opts.tags {
		tag := repo.Tag(raw)
		err := writeImage(ctx, tag, img, opts, i > 0)
		if err != nil {
			return fmt.Errorf("push %q: %w", raw, err)
		}
	}

	return nil
}

func writeImage(ctx context.Context, tag name.Tag, img remote.Taggable, opts *pushOptions, tagOnly bool) error {
	remoteOpts := []remote.Option{
		remote.WithContext(ctx),
		remote.WithAuthFromKeychain(keychain),
	}

	if opts.logger != nil && !opts.logger.IsNop() {
		var wg sync.WaitGroup
		wg.Add(1)
		defer wg.Wait()

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		ch := make(chan v1.Update, 1)
		go logProgress(ctx, opts.logger, &wg, ch, tag.TagStr())
		remoteOpts = append(remoteOpts, remote.WithProgress(ch))
	}

	if tagOnly {
		return remote.Tag(tag, img, remoteOpts...)
	}
	if image, ok := img.(v1.Image); ok {
		return remote.Write(tag, image, remoteOpts...)
	}
	if index, ok := img.(v1.ImageIndex); ok {
		return remote.WriteIndex(tag, index, remoteOpts...)
	}

	return errors.New("must be an image or image index")
}

func logProgress(ctx context.Context, l types.Logger, wg *sync.WaitGroup, ch <-chan v1.Update, tag string) {
	defer func() {
		l.Println()
		wg.Done()
	}()
	l.RePrintf("Pushing tag %s", tag)
	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-ch:
			if !ok {
				return
			}
			l.RePrintf("Pushing tag %s: %d/%d", tag, update.Complete, update.Total)
		}
	}

}
