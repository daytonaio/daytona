// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
)

func (wj *WorkspaceJob) Restart(ctx context.Context, j *models.Job) error {
	err := wj.stop(ctx, j)
	if err != nil {
		return err
	}

	return wj.start(ctx, j)
}
