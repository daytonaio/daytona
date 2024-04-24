// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const ContainerNotFoundMetadata = "{\"state\": \"container not found\"}"
const WorkspaceMetadataFormat = "{\"networkId\": \"%s\"}"

func (d *DockerClient) GetWorkspaceInfo(ws *workspace.Workspace) (*workspace.WorkspaceInfo, error) {
	workspaceInfo := &workspace.WorkspaceInfo{
		Name:             ws.Name,
		ProviderMetadata: fmt.Sprintf(WorkspaceMetadataFormat, ws.Id),
	}

	projectInfos := []*workspace.ProjectInfo{}
	for _, project := range ws.Projects {
		projectInfo, err := d.GetProjectInfo(project)
		if err != nil {
			return nil, err
		}
		projectInfos = append(projectInfos, projectInfo)
	}
	workspaceInfo.Projects = projectInfos

	return workspaceInfo, nil
}

func (d *DockerClient) GetProjectInfo(project *workspace.Project) (*workspace.ProjectInfo, error) {
	isRunning := true
	info, err := d.getContainerInfo(project)
	if err != nil {
		if client.IsErrNotFound(err) {
			isRunning = false
		} else {
			return nil, err
		}
	}

	if info == nil || info.State == nil {
		return &workspace.ProjectInfo{
			Name:             project.Name,
			IsRunning:        isRunning,
			Created:          "",
			ProviderMetadata: ContainerNotFoundMetadata,
		}, nil
	}

	projectInfo := &workspace.ProjectInfo{
		Name:      project.Name,
		IsRunning: isRunning,
		Created:   info.Created,
	}

	if info.Config != nil && info.Config.Labels != nil {
		metadata, err := json.Marshal(info.Config.Labels)
		if err != nil {
			return nil, err
		}
		projectInfo.ProviderMetadata = string(metadata)
	}

	return projectInfo, nil
}

func (d *DockerClient) getContainerInfo(project *workspace.Project) (*types.ContainerJSON, error) {
	ctx := context.Background()

	info, err := d.apiClient.ContainerInspect(ctx, d.GetProjectContainerName(project))
	if err != nil {
		return nil, err
	}

	return &info, nil
}
