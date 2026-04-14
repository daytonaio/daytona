/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"fmt"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/runner/v2/specs"
	specsgen "github.com/daytonaio/runner/pkg/runner/v2/specs/gen"
)

func (e *Executor) createSandbox(ctx context.Context, job *specsgen.Job) (any, error) {
	var payload specsgen.CreateSandboxPayload
	if err := specs.ParsePayload(job.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	_, daemonVersion, err := e.docker.Create(ctx, dto.CreateSandboxDTO{CreateSandboxPayload: &payload})
	if err != nil {
		common.ContainerOperationCount.WithLabelValues("create", string(common.PrometheusOperationStatusFailure)).Inc()
		return nil, common.FormatRecoverableError(err)
	}

	common.ContainerOperationCount.WithLabelValues("create", string(common.PrometheusOperationStatusSuccess)).Inc()

	return dto.StartSandboxResponse{
		DaemonVersion: daemonVersion,
	}, nil
}

func (e *Executor) startSandbox(ctx context.Context, job *specsgen.Job) (any, error) {
	var payload specsgen.StartSandboxPayload
	if err := specs.ParsePayload(job.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	_, daemonVersion, err := e.docker.Start(ctx, job.ResourceId, payload.AuthToken, payload.GetMetadata())
	if err != nil {
		return nil, common.FormatRecoverableError(err)
	}

	return dto.StartSandboxResponse{
		DaemonVersion: daemonVersion,
	}, nil
}

func (e *Executor) stopSandbox(ctx context.Context, job *specsgen.Job) (any, error) {
	var stopDto dto.StopSandboxDTO
	if job.Payload != nil {
		var payload specsgen.StopSandboxPayload
		if err := specs.ParsePayload(job.Payload, &payload); err == nil {
			stopDto = dto.StopSandboxDTO{StopSandboxPayload: &payload}
		}
	}

	err := e.docker.Stop(ctx, job.ResourceId, stopDto.GetForce())
	if err != nil {
		return nil, common.FormatRecoverableError(err)
	}

	return nil, nil
}

func (e *Executor) destroySandbox(ctx context.Context, job *specsgen.Job) (any, error) {
	err := e.docker.Destroy(ctx, job.ResourceId)
	if err != nil {
		common.ContainerOperationCount.WithLabelValues("destroy", string(common.PrometheusOperationStatusFailure)).Inc()
		return nil, common.FormatRecoverableError(err)
	}

	common.ContainerOperationCount.WithLabelValues("destroy", string(common.PrometheusOperationStatusSuccess)).Inc()

	return nil, nil
}

func (e *Executor) updateNetworkSettings(ctx context.Context, job *specsgen.Job) (any, error) {
	var payload specsgen.UpdateNetworkSettingsPayload
	if err := specs.ParsePayload(job.Payload, &payload); err != nil {
		return nil, common.FormatRecoverableError(fmt.Errorf("failed to unmarshal payload: %w", err))
	}

	return nil, e.docker.UpdateNetworkSettings(ctx, job.ResourceId, dto.UpdateNetworkSettingsDTO{UpdateNetworkSettingsPayload: &payload})
}

func (e *Executor) recoverSandbox(ctx context.Context, job *specsgen.Job) (any, error) {
	var payload specsgen.RecoverSandboxPayload
	if err := specs.ParsePayload(job.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	err := e.docker.RecoverSandbox(ctx, job.ResourceId, dto.RecoverSandboxDTO{RecoverSandboxPayload: &payload})
	if err != nil {
		return nil, common.FormatRecoverableError(err)
	}

	return nil, nil
}

func (e *Executor) resizeSandbox(ctx context.Context, job *specsgen.Job) (any, error) {
	var payload specsgen.ResizeSandboxPayload
	if err := specs.ParsePayload(job.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	err := e.docker.Resize(ctx, job.ResourceId, dto.ResizeSandboxDTO{ResizeSandboxPayload: &payload})
	if err != nil {
		common.ContainerOperationCount.WithLabelValues("resize", string(common.PrometheusOperationStatusFailure)).Inc()
		return nil, common.FormatRecoverableError(err)
	}

	common.ContainerOperationCount.WithLabelValues("resize", string(common.PrometheusOperationStatusSuccess)).Inc()

	return nil, nil
}
