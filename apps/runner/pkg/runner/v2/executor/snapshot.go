/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/runner/v2/specs"
	specsgen "github.com/daytonaio/runner/pkg/runner/v2/specs/gen"
)

func (e *Executor) buildSnapshot(ctx context.Context, job *specsgen.Job) (any, error) {
	var payload specsgen.BuildSnapshotPayload
	if err := specs.ParsePayload(job.Payload, &payload); err != nil {
		return nil, err
	}

	request := dto.BuildSnapshotRequestDTO{BuildSnapshotPayload: &payload}

	if err := e.docker.BuildSnapshot(ctx, request); err != nil {
		return nil, err
	}

	info, err := e.docker.GetImageInfo(ctx, request.GetSnapshot())
	if err != nil {
		return nil, err
	}

	return dto.SnapshotInfoResponse{
		Name:       request.GetSnapshot(),
		SizeGB:     float64(info.Size) / (1024 * 1024 * 1024),
		Entrypoint: info.Entrypoint,
		Cmd:        info.Cmd,
		Hash:       dto.HashWithoutPrefix(info.Hash),
	}, nil
}

func (e *Executor) pullSnapshot(ctx context.Context, job *specsgen.Job) (any, error) {
	var payload specsgen.PullSnapshotPayload
	if err := specs.ParsePayload(job.Payload, &payload); err != nil {
		return nil, err
	}

	request := dto.PullSnapshotRequestDTO{PullSnapshotPayload: &payload}

	if err := e.docker.PullSnapshot(ctx, request); err != nil {
		return nil, err
	}

	info, err := e.docker.GetImageInfo(ctx, request.GetSnapshot())
	if err != nil {
		return nil, err
	}

	return dto.SnapshotInfoResponse{
		Name:       request.GetSnapshot(),
		SizeGB:     float64(info.Size) / (1024 * 1024 * 1024),
		Entrypoint: info.Entrypoint,
		Cmd:        info.Cmd,
		Hash:       dto.HashWithoutPrefix(info.Hash),
	}, nil
}

func (e *Executor) removeSnapshot(ctx context.Context, job *specsgen.Job) (any, error) {
	return nil, e.docker.RemoveImage(ctx, job.ResourceId, true)
}

func (e *Executor) inspectSnapshotInRegistry(ctx context.Context, job *specsgen.Job) (any, error) {
	var payload specsgen.InspectSnapshotInRegistryPayload
	if err := specs.ParsePayload(job.Payload, &payload); err != nil {
		return nil, err
	}

	request := dto.InspectSnapshotInRegistryRequestDTO{InspectSnapshotInRegistryPayload: &payload}

	digest, err := e.docker.InspectImageInRegistry(ctx, request.GetSnapshot(), request.GetRegistry())
	if err != nil {
		return nil, err
	}

	return dto.SnapshotDigestResponse{
		Hash:   dto.HashWithoutPrefix(digest.Digest),
		SizeGB: float64(digest.Size) / (1024 * 1024 * 1024),
	}, nil
}
