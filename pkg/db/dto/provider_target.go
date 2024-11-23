// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import "github.com/daytonaio/daytona/pkg/provider"

type ProviderTargetDTO struct {
	Name            string  `json:"name" gorm:"primaryKey"`
	ProviderName    string  `json:"providerName"`
	ProviderLabel   *string `json:"providerLabel,omitempty"`
	ProviderVersion string  `json:"providerVersion"`
	Options         string  `json:"options"`
	IsDefault       bool    `json:"isDefault"`
}

func ToProviderTargetDTO(providerTarget *provider.ProviderTarget) ProviderTargetDTO {
	return ProviderTargetDTO{
		Name:            providerTarget.Name,
		ProviderName:    providerTarget.ProviderInfo.Name,
		ProviderLabel:   providerTarget.ProviderInfo.Label,
		ProviderVersion: providerTarget.ProviderInfo.Version,
		Options:         providerTarget.Options,
		IsDefault:       providerTarget.IsDefault,
	}
}

func ToProviderTarget(providerTargetDTO ProviderTargetDTO) *provider.ProviderTarget {
	return &provider.ProviderTarget{
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
