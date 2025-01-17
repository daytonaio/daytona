// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"errors"
	"io"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type IBuildService interface {
	List(ctx context.Context, filter *BuildFilter) ([]*BuildDTO, error)
	Find(ctx context.Context, filter *BuildFilter) (*BuildDTO, error)
	Create(ctx context.Context, createBuildDTO CreateBuildDTO) (string, error)
	Delete(ctx context.Context, filter *BuildFilter, force bool) []error

	UpdateLastJob(ctx context.Context, buildId, jobId string) error
	HandleSuccessfulRemoval(ctx context.Context, id string) error
	GetBuildLogReader(ctx context.Context, buildId string) (io.Reader, error)
	GetBuildLogWriter(ctx context.Context, buildId string) (io.WriteCloser, error)
}

type BuildDTO struct {
	models.Build
	State models.ResourceState `json:"state" validate:"required"`
} //	@name	BuildDTO

type CreateBuildDTO struct {
	WorkspaceTemplateName string            `json:"workspaceTemplateName" validate:"required"`
	Branch                string            `json:"branch" validate:"required"`
	PrebuildId            *string           `json:"prebuildId" validate:"optional"`
	EnvVars               map[string]string `json:"envVars" validate:"required"`
} // @name CreateBuildDTO

type BuildFilter struct {
	StateNames  *[]models.ResourceStateName
	ShowDeleted bool
	StoreFilter stores.BuildFilter
}

var (
	ErrBuildDeleted = errors.New("build is deleted")
)

func IsBuildDeleted(err error) bool {
	return err.Error() == ErrBuildDeleted.Error()
}
