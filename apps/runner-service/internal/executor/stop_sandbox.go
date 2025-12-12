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

func (e *Executor) stopSandbox(ctx context.Context, job *apiclient.Job) error {
	sandboxId := job.GetResourceId()
	e.log.Debug("stopping sandbox", "job_id", job.GetId(), "sandbox_id", sandboxId)

	timeout := 10
	if err := e.dockerClient.ContainerStop(ctx, sandboxId, container.StopOptions{
		Timeout: &timeout,
		Signal:  "SIGTERM",
	}); err != nil {
		return fmt.Errorf("stop container: %w", err)
	}

	e.log.Info("sandbox stopped successfully", "sandbox_id", sandboxId)
	return nil
}
