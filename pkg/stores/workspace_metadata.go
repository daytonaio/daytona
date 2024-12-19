// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type WorkspaceMetadataStore interface {
	IStore
	Find(ctx context.Context, workspaceId string) (*models.WorkspaceMetadata, error)
	Save(ctx context.Context, metadata *models.WorkspaceMetadata) error
	Delete(ctx context.Context, metadata *models.WorkspaceMetadata) error
}

var (
	ErrWorkspaceMetadataNotFound = errors.New("workspace metadata not found")
)

func IsWorkspaceMetadataNotFound(err error) bool {
	return err.Error() == ErrWorkspaceMetadataNotFound.Error()
}
