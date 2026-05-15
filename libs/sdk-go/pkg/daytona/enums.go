// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import apiclient "github.com/daytonaio/daytona/libs/api-client-go"

// Type aliases re-export api-client enum types under the daytona package so
// SDK consumers never need to import the api-client directly.

// SandboxState represents the lifecycle state of a Sandbox.
type SandboxState = apiclient.SandboxState

// SandboxListSortField selects the field used to order results from [Client.List].
type SandboxListSortField = apiclient.SandboxListSortField

// SandboxListSortDirection selects ascending or descending order for [Client.List].
type SandboxListSortDirection = apiclient.SandboxListSortDirection

// CamelCase enum constants, matching idiomatic Go naming (the underlying
// api-client uses SCREAMING_SNAKE_CASE which is non-idiomatic in Go).
const (
	SandboxStateCreating         = apiclient.SANDBOXSTATE_CREATING
	SandboxStateRestoring        = apiclient.SANDBOXSTATE_RESTORING
	SandboxStateDestroyed        = apiclient.SANDBOXSTATE_DESTROYED
	SandboxStateDestroying       = apiclient.SANDBOXSTATE_DESTROYING
	SandboxStateStarted          = apiclient.SANDBOXSTATE_STARTED
	SandboxStateStopped          = apiclient.SANDBOXSTATE_STOPPED
	SandboxStateStarting         = apiclient.SANDBOXSTATE_STARTING
	SandboxStateStopping         = apiclient.SANDBOXSTATE_STOPPING
	SandboxStateError            = apiclient.SANDBOXSTATE_ERROR
	SandboxStateBuildFailed      = apiclient.SANDBOXSTATE_BUILD_FAILED
	SandboxStatePendingBuild     = apiclient.SANDBOXSTATE_PENDING_BUILD
	SandboxStateBuildingSnapshot = apiclient.SANDBOXSTATE_BUILDING_SNAPSHOT
	SandboxStateUnknown          = apiclient.SANDBOXSTATE_UNKNOWN
	SandboxStatePullingSnapshot  = apiclient.SANDBOXSTATE_PULLING_SNAPSHOT
	SandboxStateArchived         = apiclient.SANDBOXSTATE_ARCHIVED
	SandboxStateArchiving        = apiclient.SANDBOXSTATE_ARCHIVING
	SandboxStateResizing         = apiclient.SANDBOXSTATE_RESIZING
	SandboxStateSnapshotting     = apiclient.SANDBOXSTATE_SNAPSHOTTING
	SandboxStateForking          = apiclient.SANDBOXSTATE_FORKING
)

const (
	SandboxListSortFieldName           = apiclient.SANDBOXLISTSORTFIELD_NAME
	SandboxListSortFieldCpu            = apiclient.SANDBOXLISTSORTFIELD_CPU
	SandboxListSortFieldMemoryGib      = apiclient.SANDBOXLISTSORTFIELD_MEMORY_GIB
	SandboxListSortFieldDiskGib        = apiclient.SANDBOXLISTSORTFIELD_DISK_GIB
	SandboxListSortFieldLastActivityAt = apiclient.SANDBOXLISTSORTFIELD_LAST_ACTIVITY_AT
	SandboxListSortFieldCreatedAt      = apiclient.SANDBOXLISTSORTFIELD_CREATED_AT
)

const (
	SandboxListSortDirectionAsc  = apiclient.SANDBOXLISTSORTDIRECTION_ASC
	SandboxListSortDirectionDesc = apiclient.SANDBOXLISTSORTDIRECTION_DESC
)
