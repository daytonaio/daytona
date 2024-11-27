// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type TargetMetadataFilter struct {
	Id       *string
	TargetId *string
}

type TargetMetadataStore interface {
	Find(filter *TargetMetadataFilter) (*models.TargetMetadata, error)
	Save(metadata *models.TargetMetadata) error
	Delete(metadata *models.TargetMetadata) error
}

var (
	ErrTargetMetadataNotFound = errors.New("target metadata not found")
)

func IsTargetMetadataNotFound(err error) bool {
	return err.Error() == ErrTargetMetadataNotFound.Error()
}
