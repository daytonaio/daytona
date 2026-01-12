// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

type PullSnapshotRequestDTO struct {
	Snapshot            string       `json:"snapshot" validate:"required"`
	Registry            *RegistryDTO `json:"registry,omitempty"`
	DestinationRegistry *RegistryDTO `json:"destinationRegistry,omitempty"`
	DestinationRef      *string      `json:"destinationRef,omitempty"`
} //	@name	PullSnapshotRequestDTO

type BuildSnapshotRequestDTO struct {
	Snapshot               string        `json:"snapshot,omitempty"` // Snapshot ID and tag or the build's hash
	SourceRegistries       []RegistryDTO `json:"sourceRegistries,omitempty"`
	Registry               *RegistryDTO  `json:"registry,omitempty"`
	Dockerfile             string        `json:"dockerfile" validate:"required"`
	OrganizationId         string        `json:"organizationId" validate:"required"`
	Context                []string      `json:"context"`
	PushToInternalRegistry bool          `json:"pushToInternalRegistry"`
} //	@name	BuildSnapshotRequestDTO

type TagImageRequestDTO struct {
	SourceImage string `json:"sourceImage" validate:"required"`
	TargetImage string `json:"targetImage" validate:"required"`
} //	@name	TagImageRequestDTO

type PushSnapshotRequestDTO struct {
	SandboxId    string `json:"sandboxId" validate:"required"`    // ID of the sandbox to create snapshot from
	SnapshotName string `json:"snapshotName" validate:"required"` // Name for the new snapshot (e.g., "myapp-v1.0")
} //	@name	PushSnapshotRequestDTO

type PushSnapshotResponseDTO struct {
	SnapshotName string `json:"snapshotName" example:"myapp-v1.0"`
	SnapshotPath string `json:"snapshotPath" example:"snapshots/myapp-v1.0.qcow2"`
	SizeBytes    int64  `json:"sizeBytes" example:"15032385536"`
} //	@name	PushSnapshotResponseDTO

type CreateSnapshotRequestDTO struct {
	SandboxId string `json:"sandboxId" validate:"required"` // ID of the sandbox to create snapshot from
	Name      string `json:"name" validate:"required"`      // Name for the snapshot (e.g., "myapp-v1.0")
	Live      bool   `json:"live"`                          // If true, use optimistic mode (no pause); if false (default), pause VM for consistency
} //	@name	CreateSnapshotRequestDTO

type CreateSnapshotResponseDTO struct {
	Name         string `json:"name" example:"myapp-v1.0"`
	SnapshotPath string `json:"snapshotPath" example:"snapshots/myapp-v1.0.qcow2"`
	SizeBytes    int64  `json:"sizeBytes" example:"15032385536"`
	LiveMode     bool   `json:"liveMode" example:"false"` // Which mode was used
} //	@name	CreateSnapshotResponseDTO
