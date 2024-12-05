// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
)

type AddTargetConfigDTO struct {
	Name         string              `json:"name" validate:"required"`
	ProviderInfo models.ProviderInfo `json:"providerInfo" validate:"required"`
	Options      string              `json:"options" validate:"required"`
} // @name AddTargetConfigDTO

type ITargetConfigService interface {
	Add(ctx context.Context, targetConfig AddTargetConfigDTO) (*models.TargetConfig, error)
	Find(ctx context.Context, idOrName string) (*models.TargetConfig, error)
	List(ctx context.Context) ([]*models.TargetConfig, error)
	Map(ctx context.Context) (map[string]*models.TargetConfig, error)
	Delete(ctx context.Context, targetConfigId string) error
}
