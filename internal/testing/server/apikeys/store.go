//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/apikey"
)

type InMemoryApiKeyStore struct {
	apiKeys map[string]*apikey.ApiKey
}

func NewInMemoryApiKeyStore() apikey.Store {
	return &InMemoryApiKeyStore{
		apiKeys: make(map[string]*apikey.ApiKey),
	}
}

func (s *InMemoryApiKeyStore) List() ([]*apikey.ApiKey, error) {
	apiKeys := []*apikey.ApiKey{}
	for _, a := range s.apiKeys {
		apiKeys = append(apiKeys, a)
	}

	return apiKeys, nil
}

func (s *InMemoryApiKeyStore) Find(key string) (*apikey.ApiKey, error) {
	apiKey, ok := s.apiKeys[key]
	if !ok {
		return nil, errors.New("api key not found")
	}

	return apiKey, nil
}

func (s *InMemoryApiKeyStore) FindByName(name string) (*apikey.ApiKey, error) {
	for _, a := range s.apiKeys {
		if a.Name == name {
			return a, nil
		}
	}

	return nil, nil
}

func (s *InMemoryApiKeyStore) Save(apiKey *apikey.ApiKey) error {
	s.apiKeys[apiKey.KeyHash] = apiKey
	return nil
}

func (s *InMemoryApiKeyStore) Delete(apiKey *apikey.ApiKey) error {
	delete(s.apiKeys, apiKey.KeyHash)
	return nil
}
