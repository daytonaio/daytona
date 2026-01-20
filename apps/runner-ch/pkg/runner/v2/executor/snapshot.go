/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"errors"
	"time"

	apiclient "github.com/daytonaio/apiclient"
	"github.com/daytonaio/runner-ch/pkg/api/dto"
)

func (e *Executor) buildSnapshot(ctx context.Context, job *apiclient.Job) (any, error) {
	var request dto.BuildSnapshotRequestDTO
	err := e.parsePayload(job.Payload, &request)
	if err != nil {
		return nil, err
	}

	return nil, e.chClient.BuildSnapshot(ctx, request)
}

func (e *Executor) pullSnapshot(ctx context.Context, job *apiclient.Job) (any, error) {
	var request dto.PullSnapshotRequestDTO
	err := e.parsePayload(job.Payload, &request)
	if err != nil {
		return nil, err
	}

	// Start heartbeat to keep the job alive during long-running downloads
	stopHeartbeat := e.startJobHeartbeat(ctx, job.GetId(), 2*time.Minute)
	defer stopHeartbeat()

	err = e.chClient.PullSnapshot(ctx, request)
	if err != nil {
		return nil, err
	}

	info, err := e.chClient.GetImageInfo(ctx, request.Snapshot)
	if err != nil {
		return nil, err
	}

	infoResponse := dto.SnapshotInfoResponse{
		Name:       request.Snapshot,
		SizeGB:     float64(info.Size) / (1024 * 1024 * 1024), // Convert bytes to GB
		Entrypoint: info.Entrypoint,
		Cmd:        info.Cmd,
		Hash:       dto.HashWithoutPrefix(info.Hash),
	}

	return infoResponse, nil
}

func (e *Executor) removeSnapshot(ctx context.Context, job *apiclient.Job) (any, error) {
	if job.Payload == nil || *job.Payload == "" {
		return nil, errors.New("payload is required")
	}

	return nil, e.chClient.RemoveImage(ctx, *job.Payload, true)
}

func (e *Executor) pushSnapshot(ctx context.Context, job *apiclient.Job) (any, error) {
	var request dto.PushSnapshotRequestDTO
	err := e.parsePayload(job.Payload, &request)
	if err != nil {
		return nil, err
	}

	return e.chClient.PushSnapshot(ctx, request)
}

func (e *Executor) createSandboxSnapshot(ctx context.Context, job *apiclient.Job) (any, error) {
	var request dto.CreateSnapshotRequestDTO
	err := e.parsePayload(job.Payload, &request)
	if err != nil {
		return nil, err
	}

	// Start heartbeat to keep the job alive during long-running operations
	// (disk flattening and S3 upload can take 30+ minutes for large disks)
	stopHeartbeat := e.startJobHeartbeat(ctx, job.GetId(), 2*time.Minute)
	defer stopHeartbeat()

	return e.chClient.CreateSnapshot(ctx, request)
}
