// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type RunnerMetadataStore interface {
	IStore
	List(ctx context.Context) ([]*models.RunnerMetadata, error)
	Find(ctx context.Context, runnerId string) (*models.RunnerMetadata, error)
	Save(ctx context.Context, metadata *models.RunnerMetadata) error
	Delete(ctx context.Context, metadata *models.RunnerMetadata) error
}

var (
	ErrRunnerMetadataNotFound = errors.New("runner metadata not found")
)

func IsRunnerMetadataNotFound(err error) bool {
	return err.Error() == ErrRunnerMetadataNotFound.Error()
}
