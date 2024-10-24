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

func ToTargetConfig(providerTargetDTO TargetConfigDTO) *provider.TargetConfig {
	return &provider.TargetConfig{
		Name: providerTargetDTO.Name,
		ProviderInfo: provider.ProviderInfo{
			Name:    providerTargetDTO.ProviderName,
			Label:   providerTargetDTO.ProviderLabel,
			Version: providerTargetDTO.ProviderVersion,
		},
		Options:   providerTargetDTO.Options,
		IsDefault: providerTargetDTO.IsDefault,
	}
}
