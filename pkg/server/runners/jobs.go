// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runners

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
)

func (s *RunnerService) ListRunnerJobs(ctx context.Context, runnerId string) ([]*models.Job, error) {
	return s.listJobsForRunner(ctx, runnerId)
}

func (s *RunnerService) UpdateJobState(ctx context.Context, jobId string, req services.UpdateJobStateDTO) error {
	return s.updateJobState(ctx, jobId, req)
}
