// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import (
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
)

type ApiKeyServiceConfig struct {
	ApiKeyStore stores.ApiKeyStore
}

func NewApiKeyService(config ApiKeyServiceConfig) services.IApiKeyService {
	return &ApiKeyService{
		apiKeyStore: config.ApiKeyStore,
	}
}

type ApiKeyService struct {
	apiKeyStore stores.ApiKeyStore
}
