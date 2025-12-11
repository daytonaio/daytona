/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"

	apiclient "github.com/daytonaio/apiclient"
)

func (e *Executor) destroySandbox(ctx context.Context, job *apiclient.Job) error {
	sandboxId := job.GetResourceId()
	e.log.Info("destroying sandbox", "job_id", job.GetId(), "sandbox_id", sandboxId)

	payload := job.GetPayload()
	cpu, _ := payload["cpu"].(float64)
	mem, _ := payload["mem"].(float64)

	// Force remove container
	if err := e.dockerClient.ContainerRemove(ctx, sandboxId, container.RemoveOptions{
		Force:         true,
		RemoveVolumes: false,
	}); err != nil {
		return fmt.Errorf("remove container: %w", err)
	}

	// Update allocations
	e.collector.DecrementAllocations(float32(cpu), float32(mem), 0)

	e.log.Info("sandbox destroyed successfully", "sandbox_id", sandboxId)
	return nil
}
