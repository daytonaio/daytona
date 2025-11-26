// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) ContainerInspect(ctx context.Context, containerId string) (container.InspectResponse, error) {
	return d.apiClient.ContainerInspect(ctx, containerId)
}
