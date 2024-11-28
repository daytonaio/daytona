// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"io"
	"time"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type IBuildService interface {
	Create(CreateBuildDTO) (string, error)
	Find(filter *stores.BuildFilter) (*models.Build, error)
	List(filter *stores.BuildFilter) ([]*models.Build, error)
	MarkForDeletion(filter *stores.BuildFilter, force bool) []error
	Delete(id string) error
	AwaitEmptyList(time.Duration) error
	GetBuildLogReader(buildId string) (io.Reader, error)
}

type CreateBuildDTO struct {
	WorkspaceTemplateName string            `json:"workspaceTemplateName" validate:"required"`
	Branch                string            `json:"branch" validate:"required"`
	PrebuildId            *string           `json:"prebuildId" validate:"optional"`
	EnvVars               map[string]string `json:"envVars" validate:"required"`
} // @name CreateBuildDTO
