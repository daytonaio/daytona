// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import "errors"

type ProjectConfigFilter struct {
	Name                *string
	Url                 *string
	Default             *bool
	PrebuildId          *string
	GitProviderConfigId *string
}

type PrebuildFilter struct {
	ProjectConfigName *string
	Id                *string
	Branch            *string
	CommitInterval    *int
	TriggerFiles      *[]string
}

type Store interface {
	List(filter *ProjectConfigFilter) ([]*ProjectConfig, error)
	Find(filter *ProjectConfigFilter) (*ProjectConfig, error)
	Save(projectConfig *ProjectConfig) error
	Delete(projectConfig *ProjectConfig) error
}

var (
	ErrProjectConfigNotFound = errors.New("project config not found")
	ErrPrebuildNotFound      = errors.New("prebuild not found")
)

func IsProjectConfigNotFound(err error) bool {
	return err.Error() == ErrProjectConfigNotFound.Error()
}

func IsPrebuildNotFound(err error) bool {
	return err.Error() == ErrPrebuildNotFound.Error()
}
