// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/targets"
)

type WorkspaceStore interface {
	List() ([]*models.Workspace, error)
	Find(idOrName string) (*models.Workspace, error)
	Save(workspace *models.Workspace) error
	Delete(workspace *models.Workspace) error
}

type targetStore interface {
	Find(filter *targets.TargetFilter) (*models.Target, error)
}
