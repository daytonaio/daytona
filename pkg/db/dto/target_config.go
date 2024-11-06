// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/config"
)

type TargetConfigDTO struct {
	Name         string          `json:"name" gorm:"primaryKey"`
	ProviderInfo ProviderInfoDTO `json:"providerInfo" gorm:"serializer:json"`
	Options      string          `json:"options"`
}

func ToTargetConfigDTO(targetConfig *config.TargetConfig) TargetConfigDTO {
	return TargetConfigDTO{
		Name: targetConfig.Name,
		ProviderInfo: ProviderInfoDTO{
			Name:    targetConfig.ProviderInfo.Name,
			Version: targetConfig.ProviderInfo.Version,
			Label:   targetConfig.ProviderInfo.Label,
		},
		Options: targetConfig.Options,
	}
}

func ToTargetConfig(targetConfigDTO TargetConfigDTO) *config.TargetConfig {
	return &config.TargetConfig{
		Name: targetConfigDTO.Name,
		ProviderInfo: target.ProviderInfo{
			Name:    targetConfigDTO.ProviderInfo.Name,
			Version: targetConfigDTO.ProviderInfo.Version,
			Label:   targetConfigDTO.ProviderInfo.Label,
		},
		Options: targetConfigDTO.Options,
	}
}
