// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/models"
)

type SetWorkspaceMetadata struct {
	Uptime    uint64            `json:"uptime" validate:"required"`
	GitStatus *models.GitStatus `json:"gitStatus,omitempty" validate:"optional"`
} // @name SetWorkspaceMetadata
