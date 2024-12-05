// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
)

type IApiKeyService interface {
	Generate(ctx context.Context, keyType models.ApiKeyType, name string) (string, error)
	IsWorkspaceApiKey(ctx context.Context, apiKey string) bool
	IsTargetApiKey(ctx context.Context, apiKey string) bool
	IsValidApiKey(ctx context.Context, apiKey string) bool
	ListClientKeys(ctx context.Context) ([]*models.ApiKey, error)
	Revoke(ctx context.Context, name string) error
}
