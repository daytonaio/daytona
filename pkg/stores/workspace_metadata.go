// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type WorkspaceMetadataFilter struct {
	Id          *string
	WorkspaceId *string
}

type WorkspaceMetadataStore interface {
	Find(filter *WorkspaceMetadataFilter) (*models.WorkspaceMetadata, error)
	Save(metadata *models.WorkspaceMetadata) error
	Delete(metadata *models.WorkspaceMetadata) error
}

var (
	ErrWorkspaceMetadataNotFound = errors.New("workspace metadata not found")
)

func IsWorkspaceMetadataNotFound(err error) bool {
	return err.Error() == ErrWorkspaceMetadataNotFound.Error()
}
