// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

type PullSnapshotRequestDTO struct {
	Snapshot            string       `json:"snapshot" validate:"required"` // Snapshot ref in format {orgId}/{snapshotName}
	Registry            *RegistryDTO `json:"registry,omitempty"`
	DestinationRegistry *RegistryDTO `json:"destinationRegistry,omitempty"`
	DestinationRef      *string      `json:"destinationRef,omitempty"`
} //	@name	PullSnapshotRequestDTO

type BuildSnapshotRequestDTO struct {
	Ref        string           `json:"ref" validate:"required"`
	Dockerfile string           `json:"dockerfile" validate:"required"`
	Path       string           `json:"path" validate:"required"`
	Registry   *RegistryDTO     `json:"registry,omitempty"`
	BuildArgs  *[]BuildArgument `json:"buildArgs,omitempty"`
} //	@name	BuildSnapshotRequestDTO

type BuildArgument struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
} //	@name	BuildArgument

type PushSnapshotRequestDTO struct {
	Ref      string       `json:"ref" validate:"required"`
	Registry *RegistryDTO `json:"registry,omitempty"`
} //	@name	PushSnapshotRequestDTO

type CreateSnapshotRequestDTO struct {
	SandboxId      string `json:"sandboxId" validate:"required"`      // ID of the sandbox to create snapshot from
	Name           string `json:"name" validate:"required"`           // Name for the snapshot (e.g., "myapp-v1.0")
	OrganizationId string `json:"organizationId" validate:"required"` // Organization ID to namespace the snapshot in S3
	Live           bool   `json:"live"`                               // If true, use optimistic mode (no pause); if false (default), pause VM for consistency
} //	@name	CreateSnapshotRequestDTO

type CreateSnapshotResponseDTO struct {
	Name         string `json:"name" example:"myapp-v1.0"`
	SnapshotPath string `json:"snapshotPath" example:"snapshots/myapp-v1.0.qcow2"`
	S3Path       string `json:"s3Path,omitempty" example:"s3://bucket/org123/myapp-v1.0"` // S3 path if uploaded
	SizeBytes    int64  `json:"sizeBytes" example:"15032385536"`
	LiveMode     bool   `json:"liveMode" example:"false"` // Which mode was used
} //	@name	CreateSnapshotResponseDTO

type TagImageRequestDTO struct {
	Source string `json:"source" validate:"required"`
	Target string `json:"target" validate:"required"`
} //	@name	TagImageRequestDTO

type RemoveImageRequestDTO struct {
	Ref string `json:"ref" validate:"required"`
} //	@name	RemoveImageRequestDTO

type SnapshotExistsRequestDTO struct {
	Ref string `form:"ref" validate:"required"`
} //	@name	SnapshotExistsRequestDTO

type GetSnapshotInfoRequestDTO struct {
	Ref string `form:"ref" validate:"required"`
} //	@name	GetSnapshotInfoRequestDTO

type SnapshotInfoResponseDTO struct {
	Size    int64  `json:"size"`
	Created string `json:"created"`
} //	@name	SnapshotInfoResponseDTO
