// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package models

import (
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/models/enums"
)

// SnapshotFromSandboxInfo tracks an asynchronous snapshot-from-sandbox
// capture. Name is the API-requested snapshot record name (used by the API
// poller to detect superseded captures); Info carries the resulting image
// metadata once the capture completes.
type SnapshotFromSandboxInfo struct {
	Name  string
	State enums.SnapshotFromSandboxState
	Info  *dto.SnapshotInfoResponse
	Error error
}
