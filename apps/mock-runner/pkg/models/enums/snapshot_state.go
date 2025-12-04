// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package enums

type BackupState string

const (
	BackupStateNone       BackupState = "none"
	BackupStateInProgress BackupState = "in_progress"
	BackupStateCompleted  BackupState = "completed"
	BackupStateFailed     BackupState = "failed"
)

func (s BackupState) String() string {
	return string(s)
}



