// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/target"
)

type TargetDTO struct {
	target.TargetViewDTO
	Info *target.TargetInfo `json:"info" validate:"optional"`
} //	@name	TargetDTO

type CreateTargetDTO struct {
	Id               string `json:"id" validate:"required"`
	Name             string `json:"name" validate:"required"`
	TargetConfigName string `json:"targetConfigName" validate:"required"`
} //	@name	CreateTargetDTO
