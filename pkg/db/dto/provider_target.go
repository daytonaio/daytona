// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import "github.com/daytonaio/daytona/pkg/provider"

type ProviderTargetDTO struct {
	Name            string `json:"name" gorm:"primaryKey"`
	ProviderName    string `json:"providerName"`
	ProviderVersion string `json:"providerVersion"`
	Options         string `json:"options"`
}

func ToProviderTargetDTO(providerTarget *provider.ProviderTarget) ProviderTargetDTO {
	return ProviderTargetDTO{
		Name:            providerTarget.Name,
		ProviderName:    providerTarget.ProviderInfo.Name,
		ProviderVersion: providerTarget.ProviderInfo.Version,
		Options:         providerTarget.Options,
	}
}

func ToProviderTarget(providerTargetDTO ProviderTargetDTO) *provider.ProviderTarget {
	return &provider.ProviderTarget{
		Name: providerTargetDTO.Name,
		ProviderInfo: provider.ProviderInfo{
			Name:    providerTargetDTO.ProviderName,
			Version: providerTargetDTO.ProviderVersion,
		},
		Options: providerTargetDTO.Options,
	}
}
