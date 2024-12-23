// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
)

func (pj *RunnerJob) updateProvider(ctx context.Context, j *models.Job) error {
	return pj.installProvider(ctx, j)
}
