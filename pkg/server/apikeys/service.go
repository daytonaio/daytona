// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import "github.com/daytonaio/daytona/pkg/models"

type IApiKeyService interface {
	Generate(keyType models.ApiKeyType, name string) (string, error)
	IsWorkspaceApiKey(apiKey string) bool
	IsTargetApiKey(apiKey string) bool
	IsValidApiKey(apiKey string) bool
	ListClientKeys() ([]*models.ApiKey, error)
	Revoke(name string) error
}

type ApiKeyServiceConfig struct {
	ApiKeyStore ApiKeyStore
}

func NewApiKeyService(config ApiKeyServiceConfig) IApiKeyService {
	return &ApiKeyService{
		apiKeyStore: config.ApiKeyStore,
	}
}

type ApiKeyService struct {
	apiKeyStore ApiKeyStore
}
