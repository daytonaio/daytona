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
)

func (e *Executor) createBackup(ctx context.Context, job *apiclient.Job) (any, error) {
	var createBackupDto dto.CreateBackupDTO
	err := e.parsePayload(job.Payload, &createBackupDto)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// TODO: is state cache needed?
	return nil, e.docker.CreateBackup(ctx, job.ResourceId, createBackupDto)
}
