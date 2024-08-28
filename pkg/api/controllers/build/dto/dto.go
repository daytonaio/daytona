// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

type CreateBuildDTO struct {
	ProjectConfigName string  `json:"projectConfigName" validate:"required"`
	Branch            *string `json:"branch" validate:"optional"`
	PrebuildId        *string `json:"prebuildId" validate:"optional"`
} // @name CreateBuildDTO
