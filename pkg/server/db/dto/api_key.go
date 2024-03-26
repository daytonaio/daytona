// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/types"
)

type ApiKeyDTO struct {
	KeyHash string `gorm:"primaryKey"`
	Type    types.ApiKeyType
	Name    string `gorm:"uniqueIndex"`
}

func ToApiKeyDTO(apiKey types.ApiKey) ApiKeyDTO {
	return ApiKeyDTO{
		KeyHash: apiKey.KeyHash,
		Type:    apiKey.Type,
		Name:    apiKey.Name,
	}
}

func ToApiKey(apiKeyDTO ApiKeyDTO) types.ApiKey {
	return types.ApiKey{
		KeyHash: apiKeyDTO.KeyHash,
		Type:    apiKeyDTO.Type,
		Name:    apiKeyDTO.Name,
	}
}
