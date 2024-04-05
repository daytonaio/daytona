// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

type Store interface {
	List() ([]*GitProvider, error)
	Get(id string) (*GitProvider, error)
	Set(gitProvider *GitProvider) error
	Remove(gitProvider *GitProvider) error
}
