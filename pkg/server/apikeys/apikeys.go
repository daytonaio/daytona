// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
)

func (s *ApiKeyService) ListClientKeys(ctx context.Context) ([]*models.ApiKey, error) {
	keys, err := s.apiKeyStore.List(ctx)
	if err != nil {
		return nil, err
	}

	clientKeys := []*models.ApiKey{}

	for _, key := range keys {
		if key.Type == models.ApiKeyTypeClient {
			clientKeys = append(clientKeys, key)
		}
	}

	return clientKeys, nil
}

func (s *ApiKeyService) Revoke(ctx context.Context, name string) error {
	apiKey, err := s.apiKeyStore.FindByName(ctx, name)
	if err != nil {
		return err
	}

	return s.apiKeyStore.Delete(ctx, apiKey)
}

func (s *ApiKeyService) Generate(ctx context.Context, keyType models.ApiKeyType, name string) (string, error) {
	key := s.generateRandomKey(name)

	apiKey := &models.ApiKey{
		KeyHash: s.getKeyHash(key),
		Type:    keyType,
		Name:    name,
	}

	err := s.apiKeyStore.Save(ctx, apiKey)
	if err != nil {
		return "", err
	}

	return key, nil
}
