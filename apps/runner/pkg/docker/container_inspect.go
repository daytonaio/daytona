// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"

	"github.com/containerd/errdefs"
	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) ContainerInspect(ctx context.Context, containerId string) (*container.InspectResponse, error) {
	container, err := d.apiClient.ContainerInspect(ctx, containerId)
	if err != nil {
		errWrapped := fmt.Errorf("failed to inspect sandbox container %s: %w", containerId, err)

		if errdefs.IsNotFound(err) {
			return nil, common_errors.NewNotFoundError(errWrapped)
		}

		return nil, errWrapped
	}
	return &container, nil
}
