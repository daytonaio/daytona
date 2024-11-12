// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import "github.com/daytonaio/daytona/pkg/models"

type CreateTargetConfigDTO struct {
	Name         string              `json:"name" validate:"required"`
	ProviderInfo models.ProviderInfo `json:"providerInfo" validate:"required"`
	Options      string              `json:"options" validate:"required"`
} // @name CreateTargetConfigDTO
