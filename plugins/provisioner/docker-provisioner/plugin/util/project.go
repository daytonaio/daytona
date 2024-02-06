package util

import (
	"context"

	"github.com/daytonaio/daytona/agent/workspace"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func GetContainerName(project workspace.Project) string {
	return project.Workspace.Name + "-" + project.Name
}

func GetVolumeName(project workspace.Project) string {
	return GetContainerName(project)
}

func GetContainerInfo(project workspace.Project) (*types.ContainerJSON, error) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	inspect, err := cli.ContainerInspect(ctx, GetContainerName(project))
	if err != nil {
		return nil, err
	}

	return &inspect, nil
}
