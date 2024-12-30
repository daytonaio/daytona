// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
)

type RepositoryUrl struct {
	URL string `json:"url" validate:"required"`
} // @name RepositoryUrl

type SetGitProviderConfig struct {
	Id            string                     `json:"id" validate:"optional"`
	ProviderId    string                     `json:"providerId" validate:"required"`
	Username      *string                    `json:"username,omitempty" validate:"optional"`
	Token         string                     `json:"token" validate:"required"`
	BaseApiUrl    *string                    `json:"baseApiUrl,omitempty" validate:"optional"`
	Alias         *string                    `json:"alias,omitempty" validate:"optional"`
	SigningKey    *string                    `json:"signingKey,omitempty" validate:"optional"`
	SigningMethod *gitprovider.SigningMethod `json:"signingMethod,omitempty" validate:"optional"`
} // @name SetGitProviderConfig
