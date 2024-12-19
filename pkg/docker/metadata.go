// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/docker/docker/api/types"
)

func (d *DockerClient) GetTargetProviderMetadata(t *models.Target) (string, error) {
	return "", nil
}

func (d *DockerClient) GetWorkspaceProviderMetadata(w *models.Workspace) (string, error) {
	info, err := d.getContainerInfo(w)
	if err != nil {
		return "", err
	}

	if info.Config == nil || info.Config.Labels == nil {
		return "", errors.New("container labels not found")
	}

	metadata, err := json.Marshal(info.Config.Labels)
	if err != nil {
		return "", err
	}
	return string(metadata), nil
}

func (d *DockerClient) getContainerInfo(w *models.Workspace) (*types.ContainerJSON, error) {
	ctx := context.Background()

	info, err := d.apiClient.ContainerInspect(ctx, d.GetWorkspaceContainerName(w))
	if err != nil {
		return nil, err
	}

	return &info, nil
}
