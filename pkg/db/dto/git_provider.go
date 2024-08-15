// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
)

type TokenScopeType string

const (
	TokenScopeTypeGlobal       TokenScopeType = "GLOBAL"       // Token has access to all repos and orgs
	TokenScopeTypeOrganization TokenScopeType = "ORGANIZATION" // Token is scoped to a specific organization
	TokenScopeTypeRepository   TokenScopeType = "REPOSITORY"   // Token is scoped to a specific repository
)

type GitProviderConfigDTO struct {
	ConfigId       uint           `gorm:"primaryKey;autoIncrement"`
	Id             string         `json:"id"`
	Username       string         `json:"username"`
	BaseApiUrl     *string        `json:"baseApiUrl,omitempty"`
	Token          string         `json:"token"`
	TokenIdentity  string         `json:"tokenIdentity"`
	TokenScope     string         `json:"tokenScope"`
	TokenScopeType TokenScopeType `json:"tokenScopeType"`
}

func ToGitProviderConfigDTO(gitProvider gitprovider.GitProviderConfig) GitProviderConfigDTO {
	gitProviderDTO := GitProviderConfigDTO{
		ConfigId:       gitProvider.ConfigId,
		Id:             gitProvider.Id,
		Username:       gitProvider.Username,
		BaseApiUrl:     gitProvider.BaseApiUrl,
		Token:          gitProvider.Token,
		TokenIdentity:  gitProvider.TokenIdentity,
		TokenScope:     gitProvider.TokenScope,
		TokenScopeType: TokenScopeType(gitProvider.TokenScopeType),
	}

	return gitProviderDTO
}

func ToGitProviderConfig(gitProviderDTO GitProviderConfigDTO) gitprovider.GitProviderConfig {
	return gitprovider.GitProviderConfig{
		ConfigId:       gitProviderDTO.ConfigId,
		Id:             gitProviderDTO.Id,
		Username:       gitProviderDTO.Username,
		BaseApiUrl:     gitProviderDTO.BaseApiUrl,
		Token:          gitProviderDTO.Token,
		TokenIdentity:  gitProviderDTO.TokenIdentity,
		TokenScope:     gitProviderDTO.TokenScope,
		TokenScopeType: gitprovider.TokenScopeType(gitProviderDTO.TokenScopeType),
	}
}
