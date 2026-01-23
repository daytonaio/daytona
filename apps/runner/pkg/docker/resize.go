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
	if sandboxDto.Disk != 0 {
		return fmt.Errorf("disk resize is not supported yet")
	}

	d.statesCache.SetSandboxState(ctx, sandboxId, enums.SandboxStateResizing)

	_, err := d.apiClient.ContainerUpdate(ctx, sandboxId, container.UpdateConfig{
		Resources: container.Resources{
			CPUQuota:   sandboxDto.Cpu * 100000, // Convert CPU cores to quota (1 core = 100000)
			CPUPeriod:  100000,
			Memory:     common.GBToBytes(float64(sandboxDto.Memory)),
			MemorySwap: common.GBToBytes(float64(sandboxDto.Memory)), // Set swap equal to memory to disable swap
		},
	})

	return err
}
