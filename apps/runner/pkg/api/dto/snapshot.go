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

// CreateSnapshotFromSandboxRequestDTO is the body sent by the API to ask the
// runner to commit the given sandbox container, push the resulting image to
// the supplied internal registry under the canonical
// `daytona-{hash}:daytona` tag, and return the resulting image metadata.
type CreateSnapshotFromSandboxRequestDTO struct {
	Name           string       `json:"name" validate:"required" example:"my-snapshot:latest"`
	OrganizationId string       `json:"organizationId,omitempty"`
	Registry       *RegistryDTO `json:"registry" validate:"required"`
} //	@name	CreateSnapshotFromSandboxRequest

func HashWithoutPrefix(hash string) string {
	return strings.TrimPrefix(hash, "sha256:")
}
