// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runners

import (
	"context"
)

// TODO: add lock when running interval func
// 1 second interval
const DEFAULT_JOB_POLL_INTERVAL = "*/1 * * * * *"

type IJobRunner interface {
	StartRunner(ctx context.Context) error
	CheckAndRunJobs(ctx context.Context) error
}
