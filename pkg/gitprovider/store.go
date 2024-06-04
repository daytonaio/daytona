// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import "errors"

type ConfigStore interface {
	List() ([]*GitProviderConfig, error)
	Find(id string) (*GitProviderConfig, error)
	Save(*GitProviderConfig) error
	Delete(*GitProviderConfig) error
}

var (
	ErrGitProviderNotFound = errors.New("git provider not found")
)

func IsGitProviderNotFound(err error) bool {
	return err.Error() == ErrGitProviderNotFound.Error()
}
