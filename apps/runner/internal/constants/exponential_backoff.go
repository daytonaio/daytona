// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package constants

import "time"

const (
	// Default retry configuration for Docker operations
	DEFAULT_MAX_RETRIES int           = 10
	DEFAULT_BASE_DELAY  time.Duration = 100 * time.Millisecond
	DEFAULT_MAX_DELAY   time.Duration = 5 * time.Second
)
