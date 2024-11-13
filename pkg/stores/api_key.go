// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type ApiKeyStore interface {
	List() ([]*models.ApiKey, error)
	Find(key string) (*models.ApiKey, error)
	FindByName(name string) (*models.ApiKey, error)
	Save(apiKey *models.ApiKey) error
	Delete(apiKey *models.ApiKey) error
}

var (
	ErrApiKeyNotFound = errors.New("api key not found")
)

func IsApiKeyNotFound(err error) bool {
	return err.Error() == ErrApiKeyNotFound.Error()
}
