// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func (d *DockerClient) ContainerInspect(ctx context.Context, containerId string) (container.InspectResponse, error) {
	container, err := d.apiClient.ContainerInspect(ctx, containerId)
	if err != nil {
		errWrapped := fmt.Errorf("failed to inspect sandbox container %s: %w", containerId, err)

		if client.IsErrNotFound(err) {
			return types.ContainerJSON{}, common_errors.NewNotFoundError(errWrapped)
		}

		return types.ContainerJSON{}, errWrapped
	}
	return container, nil
}
