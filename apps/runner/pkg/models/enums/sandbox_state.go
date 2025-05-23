// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package enums

type SandboxState string

const (
	SandboxStateCreating        SandboxState = "creating"
	SandboxStateRestoring       SandboxState = "restoring"
	SandboxStateDestroyed       SandboxState = "destroyed"
	SandboxStateDestroying      SandboxState = "destroying"
	SandboxStateStarted         SandboxState = "started"
	SandboxStateStopped         SandboxState = "stopped"
	SandboxStateStarting        SandboxState = "starting"
	SandboxStateStopping        SandboxState = "stopping"
	SandboxStateResizing        SandboxState = "resizing"
	SandboxStateError           SandboxState = "error"
	SandboxStateUnknown         SandboxState = "unknown"
	SandboxStatePullingSnapshot SandboxState = "pulling_snapshot"
)

func (s SandboxState) String() string {
	return string(s)
}
