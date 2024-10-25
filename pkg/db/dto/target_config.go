// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import "github.com/daytonaio/daytona/pkg/provider"

type TargetConfigDTO struct {
	Name            string  `json:"name" gorm:"primaryKey"`
	ProviderName    string  `json:"providerName"`
	ProviderLabel   *string `json:"providerLabel,omitempty"`
	ProviderVersion string  `json:"providerVersion"`
	Options         string  `json:"options"`
	IsDefault       bool    `json:"isDefault"`
}

func ToTargetConfigDTO(targetConfig *provider.TargetConfig) TargetConfigDTO {
	return TargetConfigDTO{
		Name:            targetConfig.Name,
		ProviderName:    targetConfig.ProviderInfo.Name,
		ProviderLabel:   targetConfig.ProviderInfo.Label,
		ProviderVersion: targetConfig.ProviderInfo.Version,
		Options:         targetConfig.Options,
		IsDefault:       targetConfig.IsDefault,
	}
}

func ToTargetConfig(targetConfigDTO TargetConfigDTO) *provider.TargetConfig {
	return &provider.TargetConfig{
		Name: targetConfigDTO.Name,
		ProviderInfo: provider.ProviderInfo{
			Name:    targetConfigDTO.ProviderName,
			Label:   targetConfigDTO.ProviderLabel,
			Version: targetConfigDTO.ProviderVersion,
		},
		Options:   targetConfigDTO.Options,
		IsDefault: targetConfigDTO.IsDefault,
	}
}
