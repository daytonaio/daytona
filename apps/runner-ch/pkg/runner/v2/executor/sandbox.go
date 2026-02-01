/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"fmt"

	apiclient "github.com/daytonaio/apiclient"
	"github.com/daytonaio/runner-ch/pkg/api/dto"
	"github.com/daytonaio/runner-ch/pkg/cloudhypervisor"
	"github.com/daytonaio/runner-ch/pkg/common"
)

func (e *Executor) createSandbox(ctx context.Context, job *apiclient.Job) (any, error) {
	var createSandboxDto dto.CreateSandboxDTO
	err := e.parsePayload(job.Payload, &createSandboxDto)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	_, daemonVersion, err := e.chClient.Create(ctx, createSandboxDto)
	if err != nil {
		common.ContainerOperationCount.WithLabelValues("create", string(common.PrometheusOperationStatusFailure)).Inc()
		return nil, err
	}

	common.ContainerOperationCount.WithLabelValues("create", string(common.PrometheusOperationStatusSuccess)).Inc()

	return dto.StartSandboxResponse{
		DaemonVersion: daemonVersion,
	}, nil
}

func (e *Executor) startSandbox(ctx context.Context, job *apiclient.Job) (any, error) {
	var metadata map[string]string
	err := e.parsePayload(job.Payload, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	daemonVersion, err := e.chClient.Start(ctx, job.ResourceId, metadata)
	if err != nil {
		return nil, err
	}

	return dto.StartSandboxResponse{
		DaemonVersion: daemonVersion,
	}, nil
}

func (e *Executor) stopSandbox(ctx context.Context, job *apiclient.Job) (any, error) {
	return nil, e.chClient.Stop(ctx, job.ResourceId)
}

func (e *Executor) destroySandbox(ctx context.Context, job *apiclient.Job) (any, error) {
	err := e.chClient.Destroy(ctx, job.ResourceId)
	if err != nil {
		common.ContainerOperationCount.WithLabelValues("destroy", string(common.PrometheusOperationStatusFailure)).Inc()
		return nil, err
	}

	common.ContainerOperationCount.WithLabelValues("destroy", string(common.PrometheusOperationStatusSuccess)).Inc()

	return nil, nil
}

func (e *Executor) updateNetworkSettings(ctx context.Context, job *apiclient.Job) (any, error) {
	var updateNetworkSettingsDto dto.UpdateNetworkSettingsDTO
	err := e.parsePayload(job.Payload, &updateNetworkSettingsDto)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil, e.chClient.UpdateNetworkSettings(ctx, job.ResourceId, updateNetworkSettingsDto)
}

// ForkSandboxPayload matches the payload sent by the API for FORK_SANDBOX jobs
type ForkSandboxPayload struct {
	SourceSandboxId string `json:"sourceSandboxId"`
	NewSandboxId    string `json:"newSandboxId"`
	SourceState     string `json:"sourceState"` // "started" or "stopped"
}

func (e *Executor) forkSandbox(ctx context.Context, job *apiclient.Job) (any, error) {
	var payload ForkSandboxPayload
	err := e.parsePayload(job.Payload, &payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	info, err := e.chClient.ForkVM(ctx, cloudhypervisor.ForkOptions{
		SourceSandboxId: payload.SourceSandboxId,
		NewSandboxId:    payload.NewSandboxId,
		SourceStopped:   payload.SourceState == "stopped",
	})
	if err != nil {
		common.ContainerOperationCount.WithLabelValues("fork", string(common.PrometheusOperationStatusFailure)).Inc()
		return nil, err
	}

	common.ContainerOperationCount.WithLabelValues("fork", string(common.PrometheusOperationStatusSuccess)).Inc()

	return dto.ForkSandboxResponseDTO{
		Id:       info.Id,
		State:    string(info.State),
		ParentId: payload.SourceSandboxId,
	}, nil
}

// CloneSandboxPayload matches the payload sent by the API for CLONE_SANDBOX jobs
type CloneSandboxPayload struct {
	SourceSandboxId string `json:"sourceSandboxId"`
	NewSandboxId    string `json:"newSandboxId"`
	SourceState     string `json:"sourceState"` // "started" or "stopped"
}

func (e *Executor) cloneSandbox(ctx context.Context, job *apiclient.Job) (any, error) {
	var payload CloneSandboxPayload
	err := e.parsePayload(job.Payload, &payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	info, err := e.chClient.CloneVM(ctx, cloudhypervisor.CloneOptions{
		SourceSandboxId: payload.SourceSandboxId,
		NewSandboxId:    payload.NewSandboxId,
	})
	if err != nil {
		common.ContainerOperationCount.WithLabelValues("clone", string(common.PrometheusOperationStatusFailure)).Inc()
		return nil, err
	}

	common.ContainerOperationCount.WithLabelValues("clone", string(common.PrometheusOperationStatusSuccess)).Inc()

	return dto.CloneSandboxResponseDTO{
		Id:              info.Id,
		State:           string(info.State),
		SourceSandboxId: payload.SourceSandboxId,
	}, nil
}
