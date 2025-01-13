// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import (
	"encoding/base64"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/google/uuid"
)

// returns the SHA-256 hash of a given key as a hexadecimal string.
func HashKey(key string) string {
	return util.Hash(key)
}

func GenerateRandomKey() string {
	uuid := uuid.NewString()
	return base64.RawStdEncoding.EncodeToString([]byte(uuid))
}
