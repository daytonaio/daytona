// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/stretchr/testify/mock"
)

type MockGitProviderConfigStore struct {
	mock.Mock
}

func (s *MockGitProviderConfigStore) ListConfigsForUrl(url string) ([]*gitprovider.GitProviderConfig, error) {
	args := s.Called(url)
	return args.Get(0).([]*gitprovider.GitProviderConfig), args.Error(1)
}
