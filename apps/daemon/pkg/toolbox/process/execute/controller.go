// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package execute

import "time"

type ExecuteController struct {
	terminationGracePeriod   time.Duration
	terminationCheckInterval time.Duration
}

func NewExecuteController(terminationGracePeriodSeconds, terminationCheckIntervalMilliseconds int) *ExecuteController {
	if terminationGracePeriodSeconds <= 0 {
		terminationGracePeriodSeconds = 5 // default to 5 seconds
	}

	if terminationCheckIntervalMilliseconds <= 0 {
		terminationCheckIntervalMilliseconds = 100 // default to 100 milliseconds
	}

	terminationGracePeriod := time.Duration(terminationGracePeriodSeconds) * time.Second
	terminationCheckInterval := time.Duration(terminationCheckIntervalMilliseconds) * time.Millisecond

	return &ExecuteController{
		terminationGracePeriod:   terminationGracePeriod,
		terminationCheckInterval: terminationCheckInterval,
	}
}
