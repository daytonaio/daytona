// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import (
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
)

type ApiKeyServiceConfig struct {
	ApiKeyStore       stores.ApiKeyStore
	GenerateRandomKey func(name string) string
	GetKeyHash        func(key string) string
}

func NewApiKeyService(config ApiKeyServiceConfig) services.IApiKeyService {
	return &ApiKeyService{
		apiKeyStore:       config.ApiKeyStore,
		generateRandomKey: config.GenerateRandomKey,
		getKeyHash:        config.GetKeyHash,
	}
}

type ApiKeyService struct {
	apiKeyStore       stores.ApiKeyStore
	generateRandomKey func(name string) string
	getKeyHash        func(key string) string
}
