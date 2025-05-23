// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

type PullSnapshotRequestDTO struct {
	Snapshot string       `json:"snapshot" validate:"required"`
	Registry *RegistryDTO `json:"registry,omitempty"`
} //	@name	PullSnapshotRequestDTO

type BuildSnapshotRequestDTO struct {
	Snapshot               string       `json:"snapshot,omitempty"` // Snapshot ID and tag or the build's hash
	Registry               *RegistryDTO `json:"registry,omitempty"`
	Dockerfile             string       `json:"dockerfile" validate:"required"`
	OrganizationId         string       `json:"organizationId" validate:"required"`
	Context                []string     `json:"context"`
	PushToInternalRegistry bool         `json:"pushToInternalRegistry"`
} //	@name	BuildSnapshotRequestDTO
