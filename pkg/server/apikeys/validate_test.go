// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys_test

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
)

func (s *ApiKeyServiceTestSuite) TestIsValidKey_True() {
	keyName := "api-key"

	require := s.Require()

	apiKey, err := s.apiKeyService.Create(context.TODO(), models.ApiKeyTypeWorkspace, keyName)
	require.Nil(err)

	res := s.apiKeyService.IsValidApiKey(context.TODO(), apiKey)
	require.True(res)
}

func (s *ApiKeyServiceTestSuite) TestIsValidKey_False() {
	unknownKey := "unknown"

	require := s.Require()

	res := s.apiKeyService.IsValidApiKey(context.TODO(), unknownKey)
	require.False(res)
}

func (s *ApiKeyServiceTestSuite) TestGetApiKeyType() {
	keyName := "workspaceKey"

	require := s.Require()

	apiKey, err := s.apiKeyService.Create(context.TODO(), models.ApiKeyTypeWorkspace, keyName)
	require.Nil(err)

	apiKeyType, err := s.apiKeyService.GetApiKeyType(context.TODO(), apiKey)
	require.Nil(err)
	require.Equal(models.ApiKeyTypeWorkspace, apiKeyType)
}
