// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) commitContainer(ctx context.Context, containerId, imageName string) error {
	log.Infof("Committing container %s...", containerId)

	commitResp, err := d.apiClient.ContainerCommit(ctx, containerId, container.CommitOptions{
		Reference: imageName,
		Pause:     false,
	})
	if err != nil {
		return err
	}

	log.Infof("Container %s committed successfully with image ID: %s", containerId, commitResp.ID)

	return nil
}
