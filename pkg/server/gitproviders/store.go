// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type GitProviderConfigStore interface {
	List() ([]*models.GitProviderConfig, error)
	Find(id string) (*models.GitProviderConfig, error)
	Save(*models.GitProviderConfig) error
	Delete(*models.GitProviderConfig) error
}

type WorkspaceConfigStore interface {
	Save(workspaceConfig *models.WorkspaceConfig) error
	List(gitProviderConfigId string) ([]*models.WorkspaceConfig, error)
}

var (
	ErrGitProviderConfigNotFound = errors.New("git provider config not found")
)

func IsGitProviderNotFound(err error) bool {
	return err.Error() == ErrGitProviderConfigNotFound.Error()
}
