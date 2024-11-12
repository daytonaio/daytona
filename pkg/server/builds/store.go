// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package builds

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type BuildStore interface {
	Find(filter *BuildFilter) (*models.Build, error)
	List(filter *BuildFilter) ([]*models.Build, error)
	Save(build *models.Build) error
	Delete(id string) error
}

var (
	ErrBuildNotFound = errors.New("build not found")
)

func IsBuildNotFound(err error) bool {
	return err.Error() == ErrBuildNotFound.Error()
}

type BuildFilter struct {
	Id            *string
	States        *[]models.BuildState
	PrebuildIds   *[]string
	GetNewest     *bool
	BuildConfig   *models.BuildConfig
	RepositoryUrl *string
	Branch        *string
	EnvVars       *map[string]string
}

func (f *BuildFilter) StatesToInterface() []interface{} {
	args := make([]interface{}, len(*f.States))
	for i, v := range *f.States {
		args[i] = v
	}
	return args
}

func (f *BuildFilter) PrebuildIdsToInterface() []interface{} {
	args := make([]interface{}, len(*f.PrebuildIds))
	for i, v := range *f.PrebuildIds {
		args[i] = v
	}
	return args
}
