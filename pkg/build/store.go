// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/workspace/buildconfig"
)

type Store interface {
	Find(filter *Filter) (*Build, error)
	List(filter *Filter) ([]*Build, error)
	Save(build *Build) error
	Delete(id string) error
}

var (
	ErrBuildNotFound = errors.New("build not found")
)

func IsBuildNotFound(err error) bool {
	return err.Error() == ErrBuildNotFound.Error()
}

type Filter struct {
	Id            *string
	States        *[]BuildState
	PrebuildIds   *[]string
	GetNewest     *bool
	BuildConfig   *buildconfig.BuildConfig
	RepositoryUrl *string
	Branch        *string
	EnvVars       *map[string]string
}

func (f *Filter) StatesToInterface() []interface{} {
	args := make([]interface{}, len(*f.States))
	for i, v := range *f.States {
		args[i] = v
	}
	return args
}

func (f *Filter) PrebuildIdsToInterface() []interface{} {
	args := make([]interface{}, len(*f.PrebuildIds))
	for i, v := range *f.PrebuildIds {
		args[i] = v
	}
	return args
}
