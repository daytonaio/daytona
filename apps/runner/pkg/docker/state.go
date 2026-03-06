// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

func (d *DockerClient) GetSandboxState(ctx context.Context, sandboxId string) (enums.SandboxState, error) {
	if sandboxId == "" {
		return enums.SandboxStateUnknown, nil
	}

	ct, err := d.ContainerInspect(ctx, sandboxId)
	if err != nil {
		if common_errors.IsNotFoundError(err) {
			return enums.SandboxStateDestroyed, nil
		}
		return enums.SandboxStateError, err
	}

	return d.getSandboxState(ctx, ct)
}

func (d *DockerClient) getSandboxState(ctx context.Context, ct *container.InspectResponse) (enums.SandboxState, error) {
	if ct == nil {
		return enums.SandboxStateUnknown, errors.New("invalid sandbox reference")
	}

	switch ct.State.Status {
	case container.StateCreated:
		return enums.SandboxStateCreating, nil

	case container.StateRunning:
		if d.isContainerPullingImage(ctx, ct.ID) {
			return enums.SandboxStatePullingSnapshot, nil
		}
		return enums.SandboxStateStarted, nil

	case container.StatePaused:
		return enums.SandboxStateStopped, nil

	case container.StateRestarting:
		return enums.SandboxStateStarting, nil

	case container.StateRemoving:
		return enums.SandboxStateDestroying, nil

	case container.StateExited:
		if ct.State.ExitCode == 0 || ct.State.ExitCode == 137 || ct.State.ExitCode == 143 {
			return enums.SandboxStateStopped, nil
		}
		return enums.SandboxStateError, fmt.Errorf("sandbox exited with code %d, reason: %s", ct.State.ExitCode, ct.State.Error)

	case container.StateDead:
		return enums.SandboxStateDestroyed, nil

	default:
		return enums.SandboxStateUnknown, nil
	}
}

// isContainerPullingImage checks if the container is still in image pulling phase
func (d *DockerClient) isContainerPullingImage(ctx context.Context, containerId string) bool {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       "10", // Look at last 10 lines
	}

	logs, err := d.apiClient.ContainerLogs(ctx, containerId, options)
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
