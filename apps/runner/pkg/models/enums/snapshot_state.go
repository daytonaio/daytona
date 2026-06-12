// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package enums

type BackupState string

const (
	BackupStateNone       BackupState = "NONE"
	BackupStatePending    BackupState = "PENDING"
	BackupStateInProgress BackupState = "IN_PROGRESS"
	BackupStateCompleted  BackupState = "COMPLETED"
	BackupStateFailed     BackupState = "FAILED"
)

func (s BackupState) String() string {
	return string(s)
}

type SnapshotFromSandboxState string

const (
	SnapshotFromSandboxStateNone       SnapshotFromSandboxState = "NONE"
	SnapshotFromSandboxStateInProgress SnapshotFromSandboxState = "IN_PROGRESS"
	SnapshotFromSandboxStateCompleted  SnapshotFromSandboxState = "COMPLETED"
	SnapshotFromSandboxStateFailed     SnapshotFromSandboxState = "FAILED"
)

func (s SnapshotFromSandboxState) String() string {
	return string(s)
}
