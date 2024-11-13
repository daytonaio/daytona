// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type ITargetConfigService interface {
	Delete(targetConfig *models.TargetConfig) error
	Find(filter *stores.TargetConfigFilter) (*models.TargetConfig, error)
	List(filter *stores.TargetConfigFilter) ([]*models.TargetConfig, error)
	Map() (map[string]*models.TargetConfig, error)
	Save(targetConfig *models.TargetConfig) error
}
