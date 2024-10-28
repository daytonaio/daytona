// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import "errors"

type WorkspaceConfigFilter struct {
	Name                *string
	Url                 *string
	Default             *bool
	PrebuildId          *string
	GitProviderConfigId *string
}

type PrebuildFilter struct {
	WorkspaceConfigName *string
	Id                  *string
	Branch              *string
	CommitInterval      *int
	TriggerFiles        *[]string
}

type Store interface {
	List(filter *WorkspaceConfigFilter) ([]*WorkspaceConfig, error)
	Find(filter *WorkspaceConfigFilter) (*WorkspaceConfig, error)
	Save(workspaceConfig *WorkspaceConfig) error
	Delete(workspaceConfig *WorkspaceConfig) error
}

var (
	ErrWorkspaceConfigNotFound = errors.New("workspace config not found")
	ErrPrebuildNotFound        = errors.New("prebuild not found")
)

func IsWorkspaceConfigNotFound(err error) bool {
	return err.Error() == ErrWorkspaceConfigNotFound.Error()
}

func IsPrebuildNotFound(err error) bool {
	return err.Error() == ErrPrebuildNotFound.Error()
}
