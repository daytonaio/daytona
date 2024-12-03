// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type ApiKeyStore interface {
	IStore
	List(ctx context.Context) ([]*models.ApiKey, error)
	Find(ctx context.Context, key string) (*models.ApiKey, error)
	FindByName(ctx context.Context, name string) (*models.ApiKey, error)
	Save(ctx context.Context, apiKey *models.ApiKey) error
	Delete(ctx context.Context, apiKey *models.ApiKey) error
}

var (
	ErrApiKeyNotFound = errors.New("api key not found")
)

func IsApiKeyNotFound(err error) bool {
	return err.Error() == ErrApiKeyNotFound.Error()
}
