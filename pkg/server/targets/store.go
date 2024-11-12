// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"github.com/daytonaio/daytona/pkg/models"
)

type TargetFilter struct {
	IdOrName *string
	Default  *bool
}

type TargetStore interface {
	List(filter *TargetFilter) ([]*models.Target, error)
	Find(filter *TargetFilter) (*models.Target, error)
	Save(target *models.Target) error
	Delete(target *models.Target) error
}
