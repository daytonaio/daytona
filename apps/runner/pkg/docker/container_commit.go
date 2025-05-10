// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) commitContainer(ctx context.Context, containerId, imageName string) error {
	const maxRetries = 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Infof("Committing container %s (attempt %d/%d)...", containerId, attempt, maxRetries)

		commitResp, err := d.apiClient.ContainerCommit(ctx, containerId, container.CommitOptions{
			Reference: imageName,
			Pause:     false,
		})
		if err == nil {
			log.Infof("Container %s committed successfully with image ID: %s", containerId, commitResp.ID)
			return nil
		}

		if attempt < maxRetries {
			log.Warnf("Failed to commit container %s (attempt %d/%d): %v", containerId, attempt, maxRetries, err)
			continue
		}

		return fmt.Errorf("failed to commit container after %d attempts: %w", maxRetries, err)
	}

	return nil
}
