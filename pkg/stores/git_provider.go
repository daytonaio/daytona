// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type GitProviderConfigStore interface {
	IStore
	List(ctx context.Context) ([]*models.GitProviderConfig, error)
	Find(ctx context.Context, id string) (*models.GitProviderConfig, error)
	Save(ctx context.Context, gpc *models.GitProviderConfig) error
	Delete(ctx context.Context, gpc *models.GitProviderConfig) error
}

var (
	ErrGitProviderConfigNotFound = errors.New("git provider config not found")
)

func IsGitProviderNotFound(err error) bool {
	return err.Error() == ErrGitProviderConfigNotFound.Error()
}
