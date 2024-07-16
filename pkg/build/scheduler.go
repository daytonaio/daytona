// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/daytonaio/daytona/pkg/scheduler"
)

type BuildScheduler struct {
	scheduler.AbstractScheduler
}

func NewBuildScheduler() *BuildScheduler {
	scheduler := &BuildScheduler{
		AbstractScheduler: *scheduler.NewAbstractScheduler(),
	}
	scheduler.AbstractScheduler.IScheduler = scheduler

	return scheduler
}
