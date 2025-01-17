// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/models"
)

type UpdateRunnerMetadataDTO struct {
	Uptime      uint64                `json:"uptime" validate:"required" gorm:"not null"`
	RunningJobs *uint64               `json:"runningJobs" validate:"optional" gorm:"not null"`
	Providers   []models.ProviderInfo `json:"providers" validate:"required" gorm:"serializer:json;not null"`
} // @name UpdateRunnerMetadataDTO
