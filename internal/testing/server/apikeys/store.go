//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import (
	"context"

	"github.com/daytonaio/daytona/internal/testing/common"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type InMemoryApiKeyStore struct {
	common.InMemoryStore
	apiKeys map[string]*models.ApiKey
}

func NewInMemoryApiKeyStore() stores.ApiKeyStore {
	return &InMemoryApiKeyStore{
		apiKeys: make(map[string]*models.ApiKey),
	}
}

func (s *InMemoryApiKeyStore) List(ctx context.Context) ([]*models.ApiKey, error) {
	apiKeys := []*models.ApiKey{}
	for _, a := range s.apiKeys {
		apiKeys = append(apiKeys, a)
	}

	return apiKeys, nil
}

func (s *InMemoryApiKeyStore) Find(ctx context.Context, key string) (*models.ApiKey, error) {
	apiKey, ok := s.apiKeys[key]
	if !ok {
		return nil, stores.ErrApiKeyNotFound
	}

	return apiKey, nil
}

func (s *InMemoryApiKeyStore) FindByName(ctx context.Context, name string) (*models.ApiKey, error) {
	for _, a := range s.apiKeys {
		if a.Name == name {
			return a, nil
		}
	}

	return nil, nil
}

func (s *InMemoryApiKeyStore) Save(ctx context.Context, apiKey *models.ApiKey) error {
	s.apiKeys[apiKey.KeyHash] = apiKey
	return nil
}

func (s *InMemoryApiKeyStore) Delete(ctx context.Context, apiKey *models.ApiKey) error {
	delete(s.apiKeys, apiKey.KeyHash)
	return nil
}
