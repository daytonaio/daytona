// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import "errors"

type Store interface {
	Find(hash string) (*BuildResult, error)
	List() ([]*BuildResult, error)
	Save(buildResult *BuildResult) error
	Delete(hash string) error
}

var (
	ErrBuildNotFound = errors.New("build not found")
)

func IsBuildNotFound(err error) bool {
	return err.Error() == ErrBuildNotFound.Error()
}
