// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type BuildStore interface {
	IStore
	Find(ctx context.Context, filter *BuildFilter) (*models.Build, error)
	List(ctx context.Context, filter *BuildFilter) ([]*models.Build, error)
	Save(ctx context.Context, build *models.Build) error
	Delete(ctx context.Context, id string) error
}

var (
	ErrBuildNotFound = errors.New("build not found")
)

func IsBuildNotFound(err error) bool {
	return err.Error() == ErrBuildNotFound.Error()
}

type BuildFilter struct {
	Id            *string
	PrebuildIds   *[]string
	GetNewest     *bool
	BuildConfig   *models.BuildConfig
	RepositoryUrl *string
	Branch        *string
	EnvVars       *map[string]string
}

func (f *BuildFilter) PrebuildIdsToInterface() []interface{} {
	args := make([]interface{}, len(*f.PrebuildIds))
	for i, v := range *f.PrebuildIds {
		args[i] = v
	}
	return args
}
