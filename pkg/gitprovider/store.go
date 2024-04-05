// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

type ConfigStore interface {
	List() ([]*GitProviderConfig, error)
	Find(id string) (*GitProviderConfig, error)
	Save(*GitProviderConfig) error
	Delete(*GitProviderConfig) error
}
