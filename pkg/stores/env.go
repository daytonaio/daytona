// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type EnvironmentVariableStore interface {
	IStore
	List(ctx context.Context) ([]*models.EnvironmentVariable, error)
	Save(ctx context.Context, environmentVariable *models.EnvironmentVariable) error
	Delete(ctx context.Context, key string) error
}

var (
	ErrEnvironmentVariableNotFound = errors.New("environment variable not found")
)

func IsEnvironmentVariableNotFound(err error) bool {
	return err.Error() == ErrEnvironmentVariableNotFound.Error()
}
