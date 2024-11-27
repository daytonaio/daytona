// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type WorkspaceStore interface {
	List() ([]*models.Workspace, error)
	Find(idOrName string) (*models.Workspace, error)
	Save(workspace *models.Workspace) error
	Delete(workspace *models.Workspace) error
}

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
)

func IsWorkspaceNotFound(err error) bool {
	return err.Error() == ErrWorkspaceNotFound.Error()
}
