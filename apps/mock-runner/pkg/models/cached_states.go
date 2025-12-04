// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package models

import "github.com/daytonaio/mock-runner/pkg/models/enums"

type CachedStates struct {
	SandboxState      enums.SandboxState
	BackupState       enums.BackupState
	BackupErrorReason *string
}



