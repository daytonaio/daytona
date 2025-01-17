// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
)

type IApiKeyService interface {
	ListClientKeys(ctx context.Context) ([]*ApiKeyDTO, error)
	Create(ctx context.Context, keyType models.ApiKeyType, name string) (string, error)
	Delete(ctx context.Context, name string) error

	GetApiKeyType(ctx context.Context, apiKey string) (models.ApiKeyType, error)
	GetApiKeyName(ctx context.Context, apiKey string) (string, error)
	IsValidApiKey(ctx context.Context, apiKey string) bool
}

type ApiKeyDTO struct {
	Type models.ApiKeyType `json:"type" validate:"required"`
	Name string            `json:"name" validate:"required"`
} // @name	ApiKeyDTO
