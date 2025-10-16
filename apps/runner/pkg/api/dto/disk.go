// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

type DiskDTO struct {
	DiskId    string `json:"diskId"`
	MountPath string `json:"mountPath"`
} //	@name	DiskDTO

type ArchiveDiskDTO struct {
	DiskId   string       `json:"diskId" validate:"required"`
	Registry *RegistryDTO `json:"registry,omitempty"`
} //	@name	ArchiveDiskDTO

type RestoreDiskDTO struct {
	DiskId   string       `json:"diskId" validate:"required"`
	Registry *RegistryDTO `json:"registry,omitempty"`
} //	@name	RestoreDiskDTO
