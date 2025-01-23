// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type TargetFilter struct {
	IdOrName *string
	Default  *bool
}

type TargetStore interface {
	IStore
	List(ctx context.Context, filter *TargetFilter) ([]*models.Target, error)
	Find(ctx context.Context, filter *TargetFilter) (*models.Target, error)
	Save(ctx context.Context, target *models.Target) error
	Delete(ctx context.Context, target *models.Target) error
}

var (
	ErrTargetNotFound = errors.New("target not found")
)

func IsTargetNotFound(err error) bool {
	return err.Error() == ErrTargetNotFound.Error()
}
