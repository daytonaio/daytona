/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"errors"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/runner/pkg/api/dto"
)

func (e *Executor) createCheckpoint(ctx context.Context, job *apiclient.Job) (any, error) {
	var request dto.CreateSnapshotDTO
	err := e.parsePayload(job.Payload, &request)
	if err != nil {
		return nil, err
	}

	if request.SandboxId == "" {
		request.SandboxId = job.ResourceId
	}

	return e.docker.CreateSnapshot(ctx, request)
}

func (e *Executor) removeCheckpoint(ctx context.Context, job *apiclient.Job) (any, error) {
	if job.Payload == nil || *job.Payload == "" {
		return nil, errors.New("payload is required")
	}

	return nil, e.docker.RemoveImage(ctx, *job.Payload, true)
}

func (e *Executor) inspectCheckpointInRegistry(ctx context.Context, job *apiclient.Job) (any, error) {
	var request dto.InspectSnapshotInRegistryRequestDTO
	err := e.parsePayload(job.Payload, &request)
	if err != nil {
		return nil, err
	}

	digest, err := e.docker.InspectImageInRegistry(ctx, request.Snapshot, request.Registry)
	if err != nil {
		return nil, err
	}

	return dto.SnapshotDigestResponse{
		Hash:   dto.HashWithoutPrefix(digest.Digest),
		SizeGB: float64(digest.Size) / (1024 * 1024 * 1024),
	}, nil
}
