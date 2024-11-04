// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/target"
)

type TargetDTO struct {
	Id           string `gorm:"primaryKey"`
	Name         string `json:"name" gorm:"unique"`
	TargetConfig string `json:"config"`
	ApiKey       string `json:"apiKey"`
}

func ToTargetDTO(target *target.Target) TargetDTO {
	targetDTO := TargetDTO{
		Id:           target.Id,
		Name:         target.Name,
		TargetConfig: target.TargetConfig,
		ApiKey:       target.ApiKey,
	}

	return targetDTO
}

func ToTarget(targetDTO TargetDTO) *target.Target {
	target := target.Target{
		Id:           targetDTO.Id,
		Name:         targetDTO.Name,
		TargetConfig: targetDTO.TargetConfig,
		ApiKey:       targetDTO.ApiKey,
	}

	return &target
}
