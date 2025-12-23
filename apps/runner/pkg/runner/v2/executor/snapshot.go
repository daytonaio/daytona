/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"errors"

	apiclient "github.com/daytonaio/apiclient"
	"github.com/daytonaio/runner/pkg/api/dto"
)

func (e *Executor) buildSnapshot(ctx context.Context, job *apiclient.Job) (any, error) {
	var request dto.BuildSnapshotRequestDTO
	err := e.parsePayload(job.Payload, &request)
	if err != nil {
		return nil, err
	}

	err = e.docker.BuildSnapshot(ctx, request)
	if err != nil {
		return nil, err
	}

	info, err := e.docker.GetImageInfo(ctx, request.Snapshot)
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

func (e *Executor) pullSnapshot(ctx context.Context, job *apiclient.Job) (any, error) {
	var request dto.PullSnapshotRequestDTO
	err := e.parsePayload(job.Payload, &request)
	if err != nil {
		return nil, err
	}

	err = e.docker.PullSnapshot(ctx, request)
	if err != nil {
		return nil, err
	}

	info, err := e.docker.GetImageInfo(ctx, request.Snapshot)
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

	return nil, e.docker.RemoveImage(ctx, *job.Payload, true)
}
