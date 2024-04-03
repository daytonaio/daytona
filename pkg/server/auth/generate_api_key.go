// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/types"
	"github.com/google/uuid"
)

func GenerateApiKey(keyType types.ApiKeyType, name string) (string, error) {
	key := generateRandomKey()

	apiKey := &types.ApiKey{
		KeyHash: hashKey(key),
		Type:    keyType,
		Name:    name,
	}

	err := db.SaveApiKey(apiKey)
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
