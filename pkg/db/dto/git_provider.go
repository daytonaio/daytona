// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
)

type GitProviderConfigDTO struct {
	Id         string  `gorm:"primaryKey"`
	ProviderId string  `json:"providerId"`
	Username   string  `json:"username"`
	Token      string  `json:"token"`
	BaseApiUrl *string `json:"baseApiUrl,omitempty"`
	Alias      string  `json:"alias"`
}

func ToGitProviderConfigDTO(gitProvider gitprovider.GitProviderConfig) GitProviderConfigDTO {
	gitProviderDTO := GitProviderConfigDTO{
		Id:         gitProvider.Id,
		ProviderId: gitProvider.ProviderId,
		Username:   gitProvider.Username,
		Token:      gitProvider.Token,
		BaseApiUrl: gitProvider.BaseApiUrl,
		Alias:      gitProvider.Alias,
	}

	return gitProviderDTO
}

func ToGitProviderConfig(gitProviderDTO GitProviderConfigDTO) gitprovider.GitProviderConfig {
	return gitprovider.GitProviderConfig{
		Id:         gitProviderDTO.Id,
		ProviderId: gitProviderDTO.ProviderId,
		Username:   gitProviderDTO.Username,
		Token:      gitProviderDTO.Token,
		BaseApiUrl: gitProviderDTO.BaseApiUrl,
		Alias:      gitProviderDTO.Alias,
	}
}
