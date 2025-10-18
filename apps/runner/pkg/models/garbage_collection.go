// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package models

import "time"

type SandboxCleanupInfo struct {
	ID   string
	Name string
}

type SnapshotCleanupInfo struct {
	ID        string
	Name      string
	CreatedAt time.Time
}
