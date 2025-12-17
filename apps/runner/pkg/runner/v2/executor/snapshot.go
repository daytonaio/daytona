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

func (e *Executor) buildSnapshot(ctx context.Context, job *apiclient.Job) error {
	var request dto.BuildSnapshotRequestDTO
	err := e.parsePayload(job.Payload, &request)
	if err != nil {
		return err
	}

	return e.docker.BuildSnapshot(ctx, request)
}

func (e *Executor) pullSnapshot(ctx context.Context, job *apiclient.Job) error {
	var request dto.PullSnapshotRequestDTO
	err := e.parsePayload(job.Payload, &request)
	if err != nil {
		return err
	}

	return e.docker.PullSnapshot(ctx, request)
}

func (e *Executor) removeSnapshot(ctx context.Context, job *apiclient.Job) error {
	if job.Payload == nil || *job.Payload == "" {
		return errors.New("payload is required")
	}

	return e.docker.RemoveImage(ctx, *job.Payload, true)
}
