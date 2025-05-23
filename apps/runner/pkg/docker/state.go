// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func (d *DockerClient) DeduceSandboxState(ctx context.Context, sandboxId string) (enums.SandboxState, error) {
	if sandboxId == "" {
		return enums.SandboxStateUnknown, nil
	}

	container, err := d.ContainerInspect(ctx, sandboxId)
	if err != nil {
		if client.IsErrNotFound(err) {
			return enums.SandboxStateDestroyed, nil
		}
		return enums.SandboxStateError, fmt.Errorf("failed to inspect container: %w", err)
	}

	switch container.State.Status {
	case "created":
		return enums.SandboxStateCreating, nil

	case "running":
		if d.isContainerPullingImage(container.ID) {
			return enums.SandboxStatePullingSnapshot, nil
		}
		return enums.SandboxStateStarted, nil

	case "paused":
		return enums.SandboxStateStopped, nil

	case "restarting":
		return enums.SandboxStateStarting, nil

	case "removing":
		return enums.SandboxStateDestroying, nil

	case "exited":
		if container.State.ExitCode == 0 || container.State.ExitCode == 137 || container.State.ExitCode == 143 {
			return enums.SandboxStateStopped, nil
		}

		return enums.SandboxStateError, fmt.Errorf("container exited with code %d, reason: %s", container.State.ExitCode, container.State.Error)

	case "dead":
		return enums.SandboxStateDestroyed, nil

	default:
		return enums.SandboxStateUnknown, nil
	}
}

// isContainerPullingImage checks if the container is still in image pulling phase
func (d *DockerClient) isContainerPullingImage(containerId string) bool {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       "10", // Look at last 10 lines
	}

	logs, err := d.apiClient.ContainerLogs(context.Background(), containerId, options)
	if err != nil {
		return false
	}
	defer logs.Close()

	// Read logs and check for pull messages
	buf := make([]byte, 1024)
	n, _ := logs.Read(buf)
	logContent := string(buf[:n])

	return strings.Contains(logContent, "Pulling from") ||
		strings.Contains(logContent, "Downloading") ||
		strings.Contains(logContent, "Extracting")
}
