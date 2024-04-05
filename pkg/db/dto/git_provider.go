// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
)

type GitProviderDTO struct {
	Id         string `gorm:"primaryKey"`
	Username   string
	Token      string
	BaseApiUrl string
}

func ToGitProviderDTO(gitProvider gitprovider.GitProvider) GitProviderDTO {
	return GitProviderDTO{
		Id:         gitProvider.Id,
		Username:   gitProvider.Username,
		Token:      gitProvider.Token,
		BaseApiUrl: gitProvider.BaseApiUrl,
	}
}

func ToGitProvider(gitProviderDTO GitProviderDTO) gitprovider.GitProvider {
	return gitprovider.GitProvider{
		Id:         gitProviderDTO.Id,
		Username:   gitProviderDTO.Username,
		Token:      gitProviderDTO.Token,
		BaseApiUrl: gitProviderDTO.BaseApiUrl,
	}
}
