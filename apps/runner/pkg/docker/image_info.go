// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
)

type ImageInfo struct {
	Size       int64
	Entrypoint []string
	Hash       string // Image hash/digest
}

func (d *DockerClient) GetImageInfo(ctx context.Context, imageName string) (*ImageInfo, error) {
	inspect, _, err := d.apiClient.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		return nil, err
	}

	return &ImageInfo{
		Size:       inspect.Size,
		Entrypoint: inspect.Config.Entrypoint,
		Hash:       inspect.ID,
	}, nil
}
