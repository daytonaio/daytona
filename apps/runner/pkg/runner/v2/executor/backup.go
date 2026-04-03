/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"fmt"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/runner/v2/specs"
	specsgen "github.com/daytonaio/runner/pkg/runner/v2/specs/gen"
)

func (e *Executor) createBackup(ctx context.Context, job *specsgen.Job) (any, error) {
	var payload specsgen.CreateBackupPayload
	if err := specs.ParsePayload(job.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil, e.docker.CreateBackup(ctx, job.ResourceId, dto.CreateBackupDTO{CreateBackupPayload: &payload})
}
