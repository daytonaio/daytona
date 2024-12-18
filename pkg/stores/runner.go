// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type RunnerStore interface {
	IStore
	List(ctx context.Context) ([]*models.Runner, error)
	Find(ctx context.Context, idOrName string) (*models.Runner, error)
	Save(ctx context.Context, runner *models.Runner) error
	Delete(ctx context.Context, runner *models.Runner) error
}

var (
	ErrRunnerNotFound = errors.New("runner not found")
)

func IsRunnerNotFound(err error) bool {
	return err.Error() == ErrRunnerNotFound.Error()
}
