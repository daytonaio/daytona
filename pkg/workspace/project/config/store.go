// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import "errors"

type Store interface {
	List() ([]*ProjectConfig, error)
	Find(name string) (*ProjectConfig, error)
	Save(projectConfig *ProjectConfig) error
	Delete(projectConfig *ProjectConfig) error
}

var (
	ErrProjectConfigNotFound = errors.New("project config not found")
)

func IsProjectConfigNotFound(err error) bool {
	return err.Error() == ErrProjectConfigNotFound.Error()
}
