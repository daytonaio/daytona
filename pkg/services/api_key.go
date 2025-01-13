// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
)

type IApiKeyService interface {
	Generate(ctx context.Context, keyType models.ApiKeyType, name string) (string, error)
	GetApiKeyType(ctx context.Context, apiKey string) (models.ApiKeyType, error)
	IsValidApiKey(ctx context.Context, apiKey string) bool
	ListClientKeys(ctx context.Context) ([]*ApiKeyDTO, error)
	Revoke(ctx context.Context, name string) error
	GetApiKeyName(ctx context.Context, apiKey string) (string, error)
}

type ApiKeyDTO struct {
	Type models.ApiKeyType `json:"type" validate:"required"`
	Name string            `json:"name" validate:"required"`
} // @name	ApiKeyDTO
