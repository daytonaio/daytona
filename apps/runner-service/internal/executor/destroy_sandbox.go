/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types/container"

	apiclient "github.com/daytonaio/apiclient"
)

func (e *Executor) destroySandbox(ctx context.Context, job *apiclient.Job) error {
	sandboxId := job.GetResourceId()
	e.log.Debug("destroying sandbox", "job_id", job.GetId(), "sandbox_id", sandboxId)

	payload := job.GetPayload()
	var parsedPayload map[string]interface{}
	err := json.Unmarshal([]byte(payload), &parsedPayload)
	if err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	cpu, _ := parsedPayload["cpu"].(float64)
	mem, _ := parsedPayload["mem"].(float64)

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
