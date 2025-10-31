// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"log/slog"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/errdefs"
)

func (d *DockerClient) RemoveImage(ctx context.Context, imageName string, force bool) error {
	_, err := d.apiClient.ImageRemove(ctx, imageName, image.RemoveOptions{
		Force:         force,
		PruneChildren: true,
	})
	if err != nil {
		if errdefs.IsNotFound(err) {
			slog.InfoContext(ctx, "Image already removed and not found", "imageName", imageName)
			return nil
		}
		return err
	}

	slog.InfoContext(ctx, "Image deleted successfully", "imageName", imageName)

	return nil
}
