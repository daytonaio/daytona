// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import (
	"github.com/daytonaio/daytona/internal/apikeys"
	"github.com/daytonaio/daytona/pkg/apikey"
)

func (s *ApiKeyService) ListClientKeys() ([]*apikey.ApiKey, error) {
	keys, err := s.apiKeyStore.List()
	if err != nil {
		return nil, err
	}

	clientKeys := []*apikey.ApiKey{}

	for _, key := range keys {
		if key.Type == apikey.ApiKeyTypeClient {
			clientKeys = append(clientKeys, key)
		}
	}

	return clientKeys, nil
}

func (s *ApiKeyService) Revoke(name string) error {
	apiKey, err := s.apiKeyStore.FindByName(name)
	if err != nil {
		return err
	}

	return s.apiKeyStore.Delete(apiKey)
}

func (s *ApiKeyService) Generate(keyType apikey.ApiKeyType, name string) (string, error) {
	key := apikeys.GenerateRandomKey()

	apiKey := &apikey.ApiKey{
		KeyHash: apikeys.HashKey(key),
		Type:    keyType,
		Name:    name,
	}

	err := s.apiKeyStore.Save(apiKey)
	if err != nil {
		return "", err
	}

	return key, nil
}
