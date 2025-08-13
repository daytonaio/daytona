// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"strings"
)

type ImageInfo struct {
	Size       int64
	Entrypoint []string
	Cmd        []string
	Hash       string // Image hash/digest
}

func (d *DockerClient) GetImageInfo(ctx context.Context, imageName string) (*ImageInfo, error) {
	inspect, _, err := d.apiClient.ImageInspectWithRaw(ctx, imageName)
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
