// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import "github.com/daytonaio/daytona/pkg/models"

type ApiKeyViewDTO struct {
	Type    models.ApiKeyType `json:"type" validate:"required"`
	Name    string            `json:"name" validate:"required"`
	Current bool              `json:"current" validate:"required"`
} // @name ApiKeyViewDTO
