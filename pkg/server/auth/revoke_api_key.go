// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package auth

import "github.com/daytonaio/daytona/pkg/server/db"

func RevokeApiKey(name string) error {
	apiKey, err := db.FindApiKeyByName(name)
	if err != nil {
		return err
	}

	return db.DeleteApiKey(apiKey.KeyHash)
}
