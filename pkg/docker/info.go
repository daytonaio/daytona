// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const ContainerNotFoundMetadata = "{\"state\": \"container not found\"}"
const TargetMetadataFormat = "{\"networkId\": \"%s\"}"

func (d *DockerClient) GetTargetInfo(t *models.Target) (*models.TargetInfo, error) {
	targetInfo := &models.TargetInfo{
		Name:             t.Name,
		ProviderMetadata: fmt.Sprintf(TargetMetadataFormat, t.Id),
	}

	return targetInfo, nil
}

func (d *DockerClient) GetWorkspaceInfo(w *models.Workspace) (*models.WorkspaceInfo, error) {
	isRunning := true
	info, err := d.getContainerInfo(w)
	if err != nil {
		if client.IsErrNotFound(err) {
			isRunning = false
		} else {
			return nil, err
		}
	}

	if info == nil || info.State == nil {
		return &models.WorkspaceInfo{
			Name:             w.Name,
			IsRunning:        isRunning,
			Created:          "",
			ProviderMetadata: ContainerNotFoundMetadata,
		}, nil
	}

	workspaceInfo := &models.WorkspaceInfo{
		Name:      w.Name,
		IsRunning: isRunning,
		Created:   info.Created,
	}

	if info.Config != nil && info.Config.Labels != nil {
		metadata, err := json.Marshal(info.Config.Labels)
		if err != nil {
			return nil, err
		}
		workspaceInfo.ProviderMetadata = string(metadata)
	}

	return workspaceInfo, nil
}

func (d *DockerClient) getContainerInfo(w *models.Workspace) (*types.ContainerJSON, error) {
	ctx := context.Background()

	info, err := d.apiClient.ContainerInspect(ctx, d.GetWorkspaceContainerName(w))
	if err != nil {
		return nil, err
	}

	return &info, nil
}
