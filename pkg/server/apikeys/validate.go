// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import (
	"github.com/daytonaio/daytona/internal/apikeys"
	"github.com/daytonaio/daytona/pkg/apikey"
)

func (s *ApiKeyService) IsValidApiKey(apiKey string) bool {
	keyHash := apikeys.HashKey(apiKey)

	_, err := s.apiKeyStore.Find(keyHash)
	return err == nil
}

func (s *ApiKeyService) IsProjectApiKey(apiKey string) bool {
	keyHash := apikeys.HashKey(apiKey)

	key, err := s.apiKeyStore.Find(keyHash)
	if err != nil {
		return false
	}

	if key.Type != apikey.ApiKeyTypeProject {
		return false
	}

	return true
}

func (s *ApiKeyService) IsTargetApiKey(apiKey string) bool {
	keyHash := apikeys.HashKey(apiKey)

	key, err := s.apiKeyStore.Find(keyHash)
	if err != nil {
		return false
	}

	if key.Type != apikey.ApiKeyTypeTarget {
		return false
	}

	return true
}
