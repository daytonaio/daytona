// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/types"
)

func IsValidApiKey(apiKey string) bool {
	keyHash := hashKey(apiKey)

	_, err := db.FindApiKey(keyHash)
	return err == nil
}

func IsProjectApiKey(apiKey string) bool {
	keyHash := hashKey(apiKey)

	key, err := db.FindApiKey(keyHash)
	if err != nil {
		return false
	}

	if key.Type != types.ApiKeyTypeProject {
		return false
	}

	return true
}
