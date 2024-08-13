// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/google/uuid"
)

// returns the SHA-256 hash of a given key as a hexadecimal string.
func HashKey(key string) string {
	keyHash := sha256.Sum256([]byte(key))
	return string(keyHash[:])
}

func GenerateRandomKey() string {
	uuid := uuid.NewString()
	return base64.RawStdEncoding.EncodeToString([]byte(uuid))
}

// Helper function that compares a key with a hash gotten from the API
func EqualsKeyHashFromApi(key string, keyHashFromApi string) bool {
	var keyHash string
	// We need to marshal then unmarshal the key to mimic the behavior of the API
	// Without this, the hash will be different on a byte level
	jsonString, err := json.Marshal(HashKey(key))
	if err != nil {
		return false
	}

	err = json.Unmarshal(jsonString, &keyHash)
	if err != nil {
		return false
	}

	return keyHash == keyHashFromApi
}
