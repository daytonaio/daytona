/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"

	apiclient "github.com/daytonaio/apiclient"
)

func (e *Executor) createBackup(ctx context.Context, job *apiclient.Job) error {
	e.log.Debug("Creating backup")

	// TODO: Implement actual backup creation
	// - Commit container to image
	// - Tag image with backup ID
	// - Optionally push to registry

	e.log.Info("Backup created (placeholder)")
	return nil
}
