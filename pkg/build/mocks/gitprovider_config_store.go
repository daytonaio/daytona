// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/stretchr/testify/mock"
)

type MockGitProviderConfigStore struct {
	mock.Mock
}

func (s *MockGitProviderConfigStore) ListConfigsForUrl(ctx context.Context, url string) ([]*models.GitProviderConfig, error) {
	args := s.Called(ctx, url)
	return args.Get(0).([]*models.GitProviderConfig), args.Error(1)
}
