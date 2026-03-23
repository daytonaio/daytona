// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package execute

import (
	"log/slog"
	"time"
)

type ExecuteController struct {
	logger                   *slog.Logger
	terminationGracePeriod   time.Duration
	terminationCheckInterval time.Duration
}

func NewExecuteController(logger *slog.Logger, terminationGracePeriod, terminationCheckInterval time.Duration) *ExecuteController {
	return &ExecuteController{
		logger:                   logger,
		terminationGracePeriod:   terminationGracePeriod,
		terminationCheckInterval: terminationCheckInterval,
	}
}
