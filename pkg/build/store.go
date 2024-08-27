// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"errors"
)

type BuildFilter struct {
	State *BuildState
}

type Store interface {
	Find(hash string) (*Build, error)
	List(filter *BuildFilter) ([]*Build, error)
	Save(build *Build) error
	Delete(hash string) error
}

var (
	ErrBuildNotFound = errors.New("build not found")
)

func IsBuildNotFound(err error) bool {
	return err.Error() == ErrBuildNotFound.Error()
}
