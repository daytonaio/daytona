//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
)

type InMemoryApiKeyStore struct {
	apiKeys map[string]*models.ApiKey
}

func NewInMemoryApiKeyStore() apikeys.ApiKeyStore {
	return &InMemoryApiKeyStore{
		apiKeys: make(map[string]*models.ApiKey),
	}
}

func (s *InMemoryApiKeyStore) List() ([]*models.ApiKey, error) {
	apiKeys := []*models.ApiKey{}
	for _, a := range s.apiKeys {
		apiKeys = append(apiKeys, a)
	}

	return apiKeys, nil
}

func (s *InMemoryApiKeyStore) Find(key string) (*models.ApiKey, error) {
	apiKey, ok := s.apiKeys[key]
	if !ok {
		return nil, apikeys.ErrApiKeyNotFound
	}

	return apiKey, nil
}

func (s *InMemoryApiKeyStore) FindByName(name string) (*models.ApiKey, error) {
	for _, a := range s.apiKeys {
		if a.Name == name {
			return a, nil
		}
	}

	return nil, nil
}

func (s *InMemoryApiKeyStore) Save(apiKey *models.ApiKey) error {
	s.apiKeys[apiKey.KeyHash] = apiKey
	return nil
}

func (s *InMemoryApiKeyStore) Delete(apiKey *models.ApiKey) error {
	delete(s.apiKeys, apiKey.KeyHash)
	return nil
}
