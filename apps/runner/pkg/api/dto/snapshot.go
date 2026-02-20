// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

import "strings"

type SnapshotInfoResponse struct {
	Name       string   `json:"name" example:"nginx:latest"`
	SizeGB     float64  `json:"sizeGB" example:"0.13"`
	Entrypoint []string `json:"entrypoint,omitempty" example:"[\"nginx\",\"-g\",\"daemon off;\"]"`
	Cmd        []string `json:"cmd,omitempty" example:"[\"nginx\",\"-g\",\"daemon off;\"]"`
	Hash       string   `json:"hash,omitempty" example:"a7be6198544f09a75b26e6376459b47c5b9972e7351d440e092c4faa9ea064ff"`
} //	@name	SnapshotInfoResponse

type SnapshotDigestResponse struct {
	Hash   string  `json:"hash" example:"a7be6198544f09a75b26e6376459b47c5b9972e7351d440e092c4faa9ea064ff"`
	SizeGB float64 `json:"sizeGB" example:"0.13"`
} //	@name	SnapshotDigestResponse

type InspectSnapshotInRegistryRequestDTO struct {
	Snapshot string       `json:"snapshot" validate:"required" example:"nginx:latest"`
	Registry *RegistryDTO `json:"registry,omitempty"`
} //	@name	InspectSnapshotInRegistryRequest

type CreateSnapshotDTO struct {
	SandboxId      string       `json:"sandboxId" validate:"required"`
	Name           string       `json:"name" validate:"required"`
	OrganizationId string       `json:"organizationId" validate:"required"`
	Registry       *RegistryDTO `json:"registry,omitempty"` // Registry to push the snapshot to
} //	@name	CreateSnapshotDTO

type CreateSnapshotResponseDTO struct {
	Name         string  `json:"name" example:"my-snapshot"`
	SnapshotPath string  `json:"snapshotPath" example:"registry:5000/daytona/org-123/my-snapshot:latest"`
	SizeGB       float64 `json:"sizeGB" example:"0.5"`
	SizeBytes    int64   `json:"sizeBytes" example:"536870912"`
	Hash         string  `json:"hash" example:"a7be6198544f09a75b26e6376459b47c5b9972e7351d440e092c4faa9ea064ff"`
} //	@name	CreateSnapshotResponseDTO

func HashWithoutPrefix(hash string) string {
	return strings.TrimPrefix(hash, "sha256:")
}
