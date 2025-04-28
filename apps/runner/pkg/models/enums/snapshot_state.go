// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package enums

type SnapshotState string

const (
	SnapshotStateNone       SnapshotState = "NONE"
	SnapshotStatePending    SnapshotState = "PENDING"
	SnapshotStateInProgress SnapshotState = "IN_PROGRESS"
	SnapshotStateCompleted  SnapshotState = "COMPLETED"
	SnapshotStateFailed     SnapshotState = "FAILED"
)

func (s SnapshotState) String() string {
	return string(s)
}
