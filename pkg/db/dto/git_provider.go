// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
)

type GitProviderConfigDTO struct {
	Id         string  `gorm:"primaryKey"`
	Username   string  `json:"username"`
	Token      string  `json:"token"`
	BaseApiUrl *string `json:"baseApiUrl,omitempty"`
}

func ToGitProviderConfigDTO(gitProvider gitprovider.GitProviderConfig) GitProviderConfigDTO {
	gitProviderDTO := GitProviderConfigDTO{
		Id:         gitProvider.Id,
		Username:   gitProvider.Username,
		Token:      gitProvider.Token,
		BaseApiUrl: gitProvider.BaseApiUrl,
	}

	return gitProviderDTO
}

func ToGitProviderConfig(gitProviderDTO GitProviderConfigDTO) gitprovider.GitProviderConfig {
	return gitprovider.GitProviderConfig{
		Id:         gitProviderDTO.Id,
		Username:   gitProviderDTO.Username,
		Token:      gitProviderDTO.Token,
		BaseApiUrl: gitProviderDTO.BaseApiUrl,
	}
}
