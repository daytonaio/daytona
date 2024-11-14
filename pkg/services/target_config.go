// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"github.com/daytonaio/daytona/pkg/models"
)

type AddTargetConfigDTO struct {
	Name         string              `json:"name" validate:"required"`
	ProviderInfo models.ProviderInfo `json:"providerInfo" validate:"required"`
	Options      string              `json:"options" validate:"required"`
} // @name AddTargetConfigDTO

type ITargetConfigService interface {
	Add(targetConfig AddTargetConfigDTO) (*models.TargetConfig, error)
	Find(idOrName string) (*models.TargetConfig, error)
	List() ([]*models.TargetConfig, error)
	Map() (map[string]*models.TargetConfig, error)
	Delete(targetConfigId string) error
}
