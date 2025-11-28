// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"

	cerrdefs "github.com/containerd/containerd/errdefs"
	"github.com/docker/docker/api/types/image"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) RemoveImage(ctx context.Context, imageName string, force bool) error {
	_, err := d.apiClient.ImageRemove(ctx, imageName, image.RemoveOptions{
		Force:         force,
		PruneChildren: true,
	})
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			log.Infof("Image %s already removed and not found", imageName)
			return nil
		}
		return err
	}

	log.Infof("Image %s deleted successfully", imageName)

	return nil
}
