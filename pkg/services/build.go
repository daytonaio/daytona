// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"errors"
	"io"
	"time"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type IBuildService interface {
	Create(CreateBuildDTO) (string, error)
	Find(filter *BuildFilter) (*BuildDTO, error)
	List(filter *BuildFilter) ([]*BuildDTO, error)
	Delete(filter *BuildFilter, force bool) []error
	HandleSuccessfulRemoval(id string) error
	AwaitEmptyList(time.Duration) error
	GetBuildLogReader(buildId string) (io.Reader, error)
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
