// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"errors"
	"fmt"

	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

func (d *DockerClient) GetSandboxState(ctx context.Context, sandboxId string) (enums.SandboxState, error) {
	if sandboxId == "" {
		return enums.SandboxStateUnknown, nil
	}

	if d.pullTracker.Contains(sandboxId) {
		return enums.SandboxStatePullingSnapshot, nil
	}

	ct, err := d.ContainerInspect(ctx, sandboxId)
	if err != nil {
		if common_errors.IsNotFoundError(err) {
			return enums.SandboxStateDestroyed, nil
		}
		return enums.SandboxStateError, err
	}

	return d.getSandboxState(ct)
}

func (d *DockerClient) getSandboxState(ct *container.InspectResponse) (enums.SandboxState, error) {
	if ct == nil {
		return enums.SandboxStateUnknown, errors.New("invalid sandbox reference")
	}

	switch ct.State.Status {
	case container.StateCreated:
		return enums.SandboxStateCreating, nil

	case container.StateRunning:
		return enums.SandboxStateStarted, nil

	case container.StatePaused:
		return enums.SandboxStateStopped, nil

	case container.StateRestarting:
		return enums.SandboxStateStarting, nil

	case container.StateRemoving:
		return enums.SandboxStateDestroying, nil

	case container.StateExited:
		if ct.State.ExitCode == 0 || ct.State.ExitCode == 137 || ct.State.ExitCode == 143 || ct.State.ExitCode == 255 {
			return enums.SandboxStateStopped, nil
		}
		return enums.SandboxStateError, fmt.Errorf("sandbox exited with code %d, reason: %s", ct.State.ExitCode, ct.State.Error)

	case container.StateDead:
		return enums.SandboxStateDestroyed, nil

	default:
		return enums.SandboxStateUnknown, nil
	}
}
