// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const ContainerNotFoundMetadata = "{\"state\": \"container not found\"}"
const TargetMetadataFormat = "{\"networkId\": \"%s\"}"

func (d *DockerClient) GetTargetInfo(t *target.Target) (*target.TargetInfo, error) {
	targetInfo := &target.TargetInfo{
		Name:             t.Name,
		ProviderMetadata: fmt.Sprintf(TargetMetadataFormat, t.Id),
	}

	return targetInfo, nil
}

func (d *DockerClient) GetWorkspaceInfo(w *workspace.Workspace) (*workspace.WorkspaceInfo, error) {
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
		return &workspace.WorkspaceInfo{
			Name:             w.Name,
			IsRunning:        isRunning,
			Created:          "",
			ProviderMetadata: ContainerNotFoundMetadata,
		}, nil
	}

	workspaceInfo := &workspace.WorkspaceInfo{
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

func (d *DockerClient) getContainerInfo(w *workspace.Workspace) (*types.ContainerJSON, error) {
	ctx := context.Background()

	info, err := d.apiClient.ContainerInspect(ctx, d.GetWorkspaceContainerName(w))
	if err != nil {
		return nil, err
	}

	return &info, nil
}
