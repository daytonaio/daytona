// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"

	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) Resize(ctx context.Context, sandboxId string, sandboxDto dto.ResizeSandboxDTO) error {
	// Disk resize is not supported yet
	// Value of 0 means "don't change" (minimum valid value is 1)
	if sandboxDto.Disk > 0 {
		return fmt.Errorf("disk resize is not supported yet")
	}

	// Check if there's anything to resize
	if sandboxDto.Cpu == 0 && sandboxDto.Memory == 0 {
		return nil // Nothing to resize
	}

	// Get the current state to restore after resize
	originalState, err := d.DeduceSandboxState(ctx, sandboxId)
	if err != nil {
		// Default to started if we can't deduce state
		originalState = enums.SandboxStateStarted
	}

	d.statesCache.SetSandboxState(ctx, sandboxId, enums.SandboxStateResizing)

	// Build resources with only the fields that need to change (0 = don't change)
	resources := container.Resources{}
	if sandboxDto.Cpu > 0 {
		resources.CPUQuota = sandboxDto.Cpu * 100000 // 1 core = 100000
		resources.CPUPeriod = 100000
	}
	if sandboxDto.Memory > 0 {
		resources.Memory = common.GBToBytes(float64(sandboxDto.Memory))
		resources.MemorySwap = resources.Memory // Disable swap
	}

	_, err = d.apiClient.ContainerUpdate(ctx, sandboxId, container.UpdateConfig{
		Resources: resources,
	})
	if err != nil {
		d.statesCache.SetSandboxState(ctx, sandboxId, originalState)
		return err
	}

	d.statesCache.SetSandboxState(ctx, sandboxId, originalState)

	return nil
}
