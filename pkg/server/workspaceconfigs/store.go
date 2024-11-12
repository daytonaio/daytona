// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfigs

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type WorkspaceConfigFilter struct {
	Name                *string
	Url                 *string
	Default             *bool
	PrebuildId          *string
	GitProviderConfigId *string
}

type PrebuildFilter struct {
	WorkspaceConfigName *string
	Id                  *string
	Branch              *string
	CommitInterval      *int
	TriggerFiles        *[]string
}

type WorkspaceConfigStore interface {
	List(filter *WorkspaceConfigFilter) ([]*models.WorkspaceConfig, error)
	Find(filter *WorkspaceConfigFilter) (*models.WorkspaceConfig, error)
	Save(workspaceConfig *models.WorkspaceConfig) error
	Delete(workspaceConfig *models.WorkspaceConfig) error
}

var (
	ErrWorkspaceConfigNotFound = errors.New("workspace config not found")
	ErrPrebuildNotFound        = errors.New("prebuild not found")
)

func IsWorkspaceConfigNotFound(err error) bool {
	return err.Error() == ErrWorkspaceConfigNotFound.Error()
}

func IsPrebuildNotFound(err error) bool {
	return err.Error() == ErrPrebuildNotFound.Error()
}
