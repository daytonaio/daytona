/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"fmt"

	apiclient "github.com/daytonaio/apiclient"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
)

func (e *Executor) createSandbox(ctx context.Context, job *apiclient.Job) error {
	var createSandboxDto dto.CreateSandboxDTO
	err := e.parsePayload(job.Payload, &createSandboxDto)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	_, err = e.docker.Create(ctx, createSandboxDto)
	if err != nil {
		// TODO: is this needed?
		// runner.StatesCache.SetSandboxState(ctx, createSandboxDto.Id, enums.SandboxStateError)
		common.ContainerOperationCount.WithLabelValues("create", string(common.PrometheusOperationStatusFailure)).Inc()
		return err
	}

	common.ContainerOperationCount.WithLabelValues("create", string(common.PrometheusOperationStatusSuccess)).Inc()

	return nil
}

func (e *Executor) startSandbox(ctx context.Context, job *apiclient.Job) error {
	var metadata map[string]string
	err := e.parsePayload(job.Payload, &metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return e.docker.Start(ctx, job.ResourceId, metadata)
}

func (e *Executor) stopSandbox(ctx context.Context, job *apiclient.Job) error {
	return e.docker.Stop(ctx, job.ResourceId)
}

func (e *Executor) destroySandbox(ctx context.Context, job *apiclient.Job) error {
	err := e.docker.Destroy(ctx, job.ResourceId)
	if err != nil {
		// TODO: is this needed?
		// runner.StatesCache.SetSandboxState(ctx, sandboxId, enums.SandboxStateError)
		common.ContainerOperationCount.WithLabelValues("destroy", string(common.PrometheusOperationStatusFailure)).Inc()
		return err
	}

	common.ContainerOperationCount.WithLabelValues("destroy", string(common.PrometheusOperationStatusSuccess)).Inc()

	return nil
}

func (e *Executor) updateNetworkSettings(ctx context.Context, job *apiclient.Job) error {
	var updateNetworkSettingsDto dto.UpdateNetworkSettingsDTO
	err := e.parsePayload(job.Payload, &updateNetworkSettingsDto)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return e.docker.UpdateNetworkSettings(ctx, job.ResourceId, updateNetworkSettingsDto)
}
