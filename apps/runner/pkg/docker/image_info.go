// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/daytonaio/runner/pkg/api/dto"
)

type ImageInfo struct {
	Size       int64
	Entrypoint []string
	Cmd        []string
	Hash       string // Image hash/digest
}

type ImageDigest struct {
	Digest string
	Size   int64
}

func (d *DockerClient) GetImageInfo(ctx context.Context, imageName string) (*ImageInfo, error) {
	inspect, err := d.apiClient.ImageInspect(ctx, imageName)
	if err != nil {
		return nil, err
	}

	// Extract digest from RepoDigests instead of using ID
	hash := inspect.ID // fallback to ID if no digest found
	if len(inspect.RepoDigests) > 0 {
		// RepoDigests format is like: "image@sha256:abc123..."
		for _, repoDigest := range inspect.RepoDigests {
			if strings.Contains(repoDigest, "@") {
				parts := strings.Split(repoDigest, "@")
				if len(parts) == 2 {
					hash = parts[1]
					break
				}
			}
		}
	}

	return &ImageInfo{
		Size:       inspect.Size,
		Entrypoint: inspect.Config.Entrypoint,
		Cmd:        inspect.Config.Cmd,
		Hash:       hash,
	}, nil
}

func (d *DockerClient) InspectImageInRegistry(ctx context.Context, imageName string, registry *dto.RegistryDTO) (*ImageDigest, error) {
	encodedRegistryAuth := ""
	if registry != nil {
		encodedRegistryAuth = base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "{ \"username\": \"%s\", \"password\": \"%s\" }", registry.Username, registry.Password))
	}

	digest, err := d.apiClient.DistributionInspect(ctx, imageName, encodedRegistryAuth)
	if err != nil {
		return nil, err
	}

	return &ImageDigest{
		Digest: digest.Descriptor.Digest.String(),
		Size:   digest.Descriptor.Size,
	}, nil
}
