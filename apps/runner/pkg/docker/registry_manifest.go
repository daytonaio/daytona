// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// getImageSizeFromRegistry fetches the manifest from a container registry and calculates
// the total (compressed) image size by summing all layer sizes + config size.
// This is needed because Docker's DistributionInspect only returns the manifest
// descriptor size, not the total image size.
func getImageSizeFromRegistry(ctx context.Context, imageName string, registry *dto.RegistryDTO) (int64, error) {
	ref, err := name.ParseReference(imageName)
	if err != nil {
		return 0, fmt.Errorf("failed to parse image reference: %w", err)
	}

	opts := []remote.Option{
		remote.WithContext(ctx),
		remote.WithPlatform(v1.Platform{OS: "linux", Architecture: "amd64"}),
	}

	if registry != nil && registry.HasAuth() {
		opts = append(opts, remote.WithAuth(&authn.Basic{
			Username: *registry.Username,
			Password: *registry.Password,
		}))
	}

	desc, err := remote.Get(ref, opts...)
	if err != nil {
		return 0, fmt.Errorf("failed to get image descriptor from registry: %w", err)
	}

	img, err := desc.Image()
	if err != nil {
		return 0, fmt.Errorf("failed to resolve image from descriptor: %w", err)
	}

	manifest, err := img.Manifest()
	if err != nil {
		return 0, fmt.Errorf("failed to get image manifest: %w", err)
	}

	var totalSize int64
	totalSize += manifest.Config.Size
	for _, layer := range manifest.Layers {
		totalSize += layer.Size
	}

	if totalSize == 0 {
		return 0, fmt.Errorf("manifest reported zero total size for %s", imageName)
	}

	return totalSize, nil
}
