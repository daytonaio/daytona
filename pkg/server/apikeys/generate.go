// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/google/uuid"
)

func (s *ApiKeyService) Generate(keyType apikey.ApiKeyType, name string) (string, error) {
	key := generateRandomKey()

	apiKey := &apikey.ApiKey{
		KeyHash: hashKey(key),
		Type:    keyType,
		Name:    name,
	}

	err := s.apiKeyStore.Save(apiKey)
	if err != nil {
		return "", err
	}

	return key, nil
}

func hashKey(key string) string {
	keyHash := sha256.Sum256([]byte(key))
	return string(keyHash[:])
}

func generateRandomKey() string {
	uuid := uuid.NewString()
	return base64.RawStdEncoding.EncodeToString([]byte(uuid))
}
