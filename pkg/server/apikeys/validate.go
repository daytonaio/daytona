// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import (
	"context"

	"github.com/daytonaio/daytona/internal/apikeys"
	"github.com/daytonaio/daytona/pkg/models"
)

func (s *ApiKeyService) IsValidApiKey(ctx context.Context, apiKey string) bool {
	keyHash := apikeys.HashKey(apiKey)

	_, err := s.apiKeyStore.Find(ctx, keyHash)
	return err == nil
}

func (s *ApiKeyService) GetApiKeyType(ctx context.Context, apiKey string) (models.ApiKeyType, error) {
	keyHash := apikeys.HashKey(apiKey)

	key, err := s.apiKeyStore.Find(ctx, keyHash)
	if err != nil {
		return models.ApiKeyTypeClient, err
	}

	return key.Type, nil
}

func (s *ApiKeyService) IsTargetApiKey(ctx context.Context, apiKey string) bool {
	keyHash := apikeys.HashKey(apiKey)

	key, err := s.apiKeyStore.Find(ctx, keyHash)
	if err != nil {
		return false
	}

	if key.Type != models.ApiKeyTypeTarget {
		return false
	}

	return true
}
