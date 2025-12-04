// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

type CreateBackupDTO struct {
	Registry RegistryDTO `json:"registry" validate:"required"`
	Snapshot string      `json:"snapshot" validate:"required"`
} //	@name	CreateBackupDTO
