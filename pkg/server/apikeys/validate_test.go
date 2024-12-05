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

	apiKey, err := s.apiKeyService.Generate(context.TODO(), models.ApiKeyTypeWorkspace, keyName)
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

func (s *ApiKeyServiceTestSuite) TestIsWorkspaceApiKey_True() {
	keyName := "workspaceKey"

	require := s.Require()

	apiKey, err := s.apiKeyService.Generate(context.TODO(), models.ApiKeyTypeWorkspace, keyName)
	require.Nil(err)

	res := s.apiKeyService.IsWorkspaceApiKey(context.TODO(), apiKey)
	require.True(res)
}

func (s *ApiKeyServiceTestSuite) TestIsWorkspaceApiKey_False() {
	keyName := "clientKey"

	require := s.Require()

	apiKey, err := s.apiKeyService.Generate(context.TODO(), models.ApiKeyTypeClient, keyName)
	require.Nil(err)

	res := s.apiKeyService.IsWorkspaceApiKey(context.TODO(), apiKey)
	require.False(res)
}

func (s *ApiKeyServiceTestSuite) TestIsIsTargetApiKey_True() {
	keyName := "targetKey"

	require := s.Require()

	apiKey, err := s.apiKeyService.Generate(context.TODO(), models.ApiKeyTypeTarget, keyName)
	require.Nil(err)

	res := s.apiKeyService.IsTargetApiKey(context.TODO(), apiKey)
	require.True(res)
}

func (s *ApiKeyServiceTestSuite) TestIsTargetApiKey_False() {
	keyName := "clientKey"

	require := s.Require()

	apiKey, err := s.apiKeyService.Generate(context.TODO(), models.ApiKeyTypeClient, keyName)
	require.Nil(err)

	res := s.apiKeyService.IsTargetApiKey(context.TODO(), apiKey)
	require.False(res)
}
