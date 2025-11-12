// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package models

import (
	"github.com/daytonaio/runner/pkg/models/enums"
)

type SandboxInfo struct {
	SandboxState      enums.SandboxState
	BackupState       enums.BackupState
	BackupErrorReason *string
}
