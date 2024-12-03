// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type WorkspaceStore interface {
	IStore
	List(ctx context.Context) ([]*models.Workspace, error)
	Find(ctx context.Context, idOrName string) (*models.Workspace, error)
	Save(ctx context.Context, workspace *models.Workspace) error
	Delete(ctx context.Context, workspace *models.Workspace) error
}

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
)

func IsWorkspaceNotFound(err error) bool {
	return err.Error() == ErrWorkspaceNotFound.Error()
}
