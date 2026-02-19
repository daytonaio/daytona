/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"fmt"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
)

func (e *Executor) createSandbox(ctx context.Context, job *apiclient.Job) (any, error) {
	var createSandboxDto dto.CreateSandboxDTO
	err := e.parsePayload(job.Payload, &createSandboxDto)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	_, daemonVersion, err := e.docker.Create(ctx, createSandboxDto)
	if err != nil {
		// TODO: is this needed?
		// runner.StatesCache.SetSandboxState(ctx, createSandboxDto.Id, enums.SandboxStateError)
		common.ContainerOperationCount.WithLabelValues("create", string(common.PrometheusOperationStatusFailure)).Inc()
		return nil, common.FormatRecoverableError(err)
	}

	common.ContainerOperationCount.WithLabelValues("create", string(common.PrometheusOperationStatusSuccess)).Inc()

	return dto.StartSandboxResponse{
		DaemonVersion: daemonVersion,
	}, nil
}

func (e *Executor) startSandbox(ctx context.Context, job *apiclient.Job) (any, error) {
	var payload StartSandboxPayload
	err := e.parsePayload(job.Payload, &payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	daemonVersion, err := e.docker.Start(ctx, job.ResourceId, payload.AuthToken, payload.Metadata)
	if err != nil {
		return nil, common.FormatRecoverableError(err)
	}

	return dto.StartSandboxResponse{
		DaemonVersion: daemonVersion,
	}, nil
}

func (e *Executor) stopSandbox(ctx context.Context, job *apiclient.Job) (any, error) {
	err := e.docker.Stop(ctx, job.ResourceId)
	if err != nil {
		return nil, common.FormatRecoverableError(err)
	}

	return nil, nil
}

func (e *Executor) destroySandbox(ctx context.Context, job *apiclient.Job) (any, error) {
	err := e.docker.Destroy(ctx, job.ResourceId)
	if err != nil {
		// TODO: is this needed?
		// runner.StatesCache.SetSandboxState(ctx, sandboxId, enums.SandboxStateError)
		common.ContainerOperationCount.WithLabelValues("destroy", string(common.PrometheusOperationStatusFailure)).Inc()
		return nil, common.FormatRecoverableError(err)
	}

	common.ContainerOperationCount.WithLabelValues("destroy", string(common.PrometheusOperationStatusSuccess)).Inc()

	return nil, nil
}

func (e *Executor) updateNetworkSettings(ctx context.Context, job *apiclient.Job) (any, error) {
	var updateNetworkSettingsDto dto.UpdateNetworkSettingsDTO
	err := e.parsePayload(job.Payload, &updateNetworkSettingsDto)
	if err != nil {
		return nil, common.FormatRecoverableError(fmt.Errorf("failed to unmarshal payload: %w", err))
	}

	return nil, e.docker.UpdateNetworkSettings(ctx, job.ResourceId, updateNetworkSettingsDto)
}

func (e *Executor) recoverSandbox(ctx context.Context, job *apiclient.Job) (any, error) {
	var recoverSandboxDto dto.RecoverSandboxDTO
	err := e.parsePayload(job.Payload, &recoverSandboxDto)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	err = e.docker.RecoverSandbox(ctx, job.ResourceId, recoverSandboxDto)
	if err != nil {
		return nil, common.FormatRecoverableError(err)
	}

	return nil, nil
}

func (e *Executor) resizeSandbox(ctx context.Context, job *apiclient.Job) (any, error) {
	var resizeSandboxDto dto.ResizeSandboxDTO
	err := e.parsePayload(job.Payload, &resizeSandboxDto)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	err = e.docker.Resize(ctx, job.ResourceId, resizeSandboxDto)
	if err != nil {
		common.ContainerOperationCount.WithLabelValues("resize", string(common.PrometheusOperationStatusFailure)).Inc()
		return nil, common.FormatRecoverableError(err)
	}

	common.ContainerOperationCount.WithLabelValues("resize", string(common.PrometheusOperationStatusSuccess)).Inc()

	return nil, nil
}
