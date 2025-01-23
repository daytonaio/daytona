// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type TargetConfigStore interface {
	IStore
	List(ctx context.Context, allowDeleted bool) ([]*models.TargetConfig, error)
	Find(ctx context.Context, idOrName string, allowDeleted bool) (*models.TargetConfig, error)
	Save(ctx context.Context, targetConfig *models.TargetConfig) error
}

var (
	ErrTargetConfigNotFound      = errors.New("target config not found")
	ErrTargetConfigAlreadyExists = errors.New("target config with the same name already exists")
)

func IsTargetConfigNotFound(err error) bool {
	return err.Error() == ErrTargetConfigNotFound.Error()
}
