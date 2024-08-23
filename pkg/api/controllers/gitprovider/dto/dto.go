// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import "github.com/daytonaio/daytona/pkg/gitprovider"

type RepositoryUrl struct {
	URL string `json:"url" validate:"required"`
} // @name RepositoryUrl

type SetGitProviderConfig struct {
	Id             string                     `json:"id" validate:"required"`
	Username       *string                    `json:"username" validate:"optional"`
	Token          string                     `json:"token" validate:"required"`
	BaseApiUrl     *string                    `json:"baseApiUrl,omitempty" validate:"optional"`
	TokenIdentity  string                     `json:"tokenIdentity"`
	TokenScope     string                     `json:"tokenScope"`
	TokenScopeType gitprovider.TokenScopeType `json:"tokenScopeType"`
} // @name SetGitProviderConfig
