// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import "github.com/daytonaio/daytona/pkg/apikey"

type IApiKeyService interface {
	Generate(keyType apikey.ApiKeyType, name string) (string, error)
	IsProjectApiKey(apiKey string) bool
	IsValidApiKey(apiKey string) bool
	ListClientKeys() ([]*apikey.ApiKey, error)
	Revoke(name string) error
}

type ApiKeyServiceConfig struct {
	ApiKeyStore apikey.Store
}

func NewApiKeyService(config ApiKeyServiceConfig) IApiKeyService {
	return &ApiKeyService{
		apiKeyStore: config.ApiKeyStore,
	}
}

type ApiKeyService struct {
	apiKeyStore apikey.Store
}
