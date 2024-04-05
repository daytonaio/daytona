// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import "github.com/daytonaio/daytona/pkg/apikey"

type ApiKeyServiceConfig struct {
	ApiKeyStore apikey.Store
}

func NewApiKeyService(config ApiKeyServiceConfig) *ApiKeyService {
	return &ApiKeyService{
		apiKeyStore: config.ApiKeyStore,
	}
}

type ApiKeyService struct {
	apiKeyStore apikey.Store
}
