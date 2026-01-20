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
)

func (e *Executor) createBackup(ctx context.Context, job *apiclient.Job) (any, error) {
	var createBackupDto dto.CreateBackupDTO
	err := e.parsePayload(job.Payload, &createBackupDto)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil, e.chClient.CreateBackup(ctx, job.ResourceId, createBackupDto)
}
