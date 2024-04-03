// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package auth

import "github.com/daytonaio/daytona/pkg/server/db"

func IsValidApiKey(apiKey string) bool {
	keyHash := hashKey(apiKey)

	_, err := db.FindApiKey(keyHash)
	return err == nil
}
