package util

import (
	"context"
	"time"

	"github.com/daytonaio/daytona/agent/workspace"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func StopContainer(project workspace.Project) error {
	containerName := GetContainerName(project)
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	err = cli.ContainerStop(ctx, containerName, container.StopOptions{})
	if err != nil {
		return err
	}

	//	TODO: timeout
	for {
		inspect, err := cli.ContainerInspect(ctx, containerName)
		if err != nil {
			return err
		}

		if !inspect.State.Running {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}
