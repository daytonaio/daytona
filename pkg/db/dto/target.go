// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/target"
)

type TargetDTO struct {
	Id           string          `gorm:"primaryKey"`
	Name         string          `json:"name" gorm:"unique"`
	ProviderInfo ProviderInfoDTO `json:"providerInfo" gorm:"serializer:json"`
	Options      string          `json:"options"`
	ApiKey       string          `json:"apiKey"`
	IsDefault    bool            `json:"isDefault"`
	Workspaces   []WorkspaceDTO  `gorm:"foreignKey:TargetId;references:Id"`
}

type ProviderInfoDTO struct {
	Name    string  `json:"name" validate:"required"`
	Version string  `json:"version" validate:"required"`
	Label   *string `json:"label" validate:"optional"`
}

func ToTargetDTO(target *target.Target) TargetDTO {
	targetDTO := TargetDTO{
		Id:   target.Id,
		Name: target.Name,
		ProviderInfo: ProviderInfoDTO{
			Name:    target.ProviderInfo.Name,
			Version: target.ProviderInfo.Version,
			Label:   target.ProviderInfo.Label,
		},
		Options:   target.Options,
		ApiKey:    target.ApiKey,
		IsDefault: target.IsDefault,
	}

	return targetDTO
}

func ToTarget(targetDTO TargetDTO) *target.Target {
	target := target.Target{
		Id:   targetDTO.Id,
		Name: targetDTO.Name,
		ProviderInfo: target.ProviderInfo{
			Name:    targetDTO.ProviderInfo.Name,
			Version: targetDTO.ProviderInfo.Version,
			Label:   targetDTO.ProviderInfo.Label,
		},
		Options:   targetDTO.Options,
		IsDefault: targetDTO.IsDefault,
		ApiKey:    targetDTO.ApiKey,
	}

	return &target
}

func ToTargetViewDTO(targetDto TargetDTO) *target.TargetViewDTO {
	return &target.TargetViewDTO{
		Target:         *ToTarget(targetDto),
		WorkspaceCount: len(targetDto.Workspaces),
	}
}

func ViewToTarget(targetViewDTO *target.TargetViewDTO) *target.Target {
	return &target.Target{
		Id:           targetViewDTO.Id,
		Name:         targetViewDTO.Name,
		ProviderInfo: targetViewDTO.ProviderInfo,
		Options:      targetViewDTO.Options,
		ApiKey:       targetViewDTO.ApiKey,
		EnvVars:      targetViewDTO.EnvVars,
		IsDefault:    targetViewDTO.IsDefault,
	}
}
