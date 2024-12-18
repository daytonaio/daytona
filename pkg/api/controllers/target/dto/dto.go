// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

type SetTargetMetadata struct {
	Uptime uint64 `json:"uptime" validate:"required"`
} // @name SetTargetMetadata

type UpdateTargetProviderMetadataDTO struct {
	Metadata string `json:"metadata" validate:"required"`
} // @name UpdateTargetProviderMetadataDTO
