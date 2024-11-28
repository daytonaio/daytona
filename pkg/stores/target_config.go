// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type TargetConfigStore interface {
	List(allowDeleted bool) ([]*models.TargetConfig, error)
	Find(idOrName string, allowDeleted bool) (*models.TargetConfig, error)
	Save(targetConfig *models.TargetConfig) error
}

var (
	ErrTargetConfigNotFound      = errors.New("target config not found")
	ErrTargetConfigAlreadyExists = errors.New("target config with the same name already exists")
)

func IsTargetConfigNotFound(err error) bool {
	return err.Error() == ErrTargetConfigNotFound.Error()
}
