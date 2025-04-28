// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/models/enums"

	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) Resize(ctx context.Context, sandboxId string, sandboxDto dto.ResizeSandboxDTO) error {
	d.cache.SetSandboxState(ctx, sandboxId, enums.SandboxStateResizing)

	_, err := d.apiClient.ContainerUpdate(ctx, sandboxId, container.UpdateConfig{
		Resources: container.Resources{
			CPUQuota:   sandboxDto.Cpu * 100000, // Convert CPU cores to quota (1 core = 100000)
			CPUPeriod:  100000,
			Memory:     sandboxDto.Memory * 1024 * 1024 * 1024, // Convert GB to bytes
			MemorySwap: sandboxDto.Memory * 1024 * 1024 * 1024, // Set swap equal to memory to disable swap
		},
	})

	return err
}
