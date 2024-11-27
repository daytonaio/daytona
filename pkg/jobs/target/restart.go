// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
)

func (tj *TargetJob) restart(ctx context.Context, j *models.Job) error {
	err := tj.stop(ctx, j)
	if err != nil {
		return err
	}

	return tj.start(ctx, j)
}
