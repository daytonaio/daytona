// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package enums

type DiskState string

const (
	DiskStateFresh     DiskState = "fresh"
	DiskStatePulling   DiskState = "pulling"
	DiskStateReady     DiskState = "ready"
	DiskStateAttached  DiskState = "attached"
	DiskStateDetached  DiskState = "detached"
	DiskStateUploading DiskState = "uploading"
	DiskStateStored    DiskState = "stored"
)
