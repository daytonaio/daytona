// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys_test

import "github.com/daytonaio/daytona/pkg/apikey"

func (s *ApiKeyServiceTestSuite) TestIsValidKey_True() {
	keyName := "api-key"

	require := s.Require()

	apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeWorkspace, keyName)
	require.Nil(err)

	res := s.apiKeyService.IsValidApiKey(apiKey)
	require.True(res)
}

func (s *ApiKeyServiceTestSuite) TestIsValidKey_False() {
	unknownKey := "unknown"

	require := s.Require()

	res := s.apiKeyService.IsValidApiKey(unknownKey)
	require.False(res)
}

func (s *ApiKeyServiceTestSuite) TestIsWorkspaceApiKey_True() {
	keyName := "workspaceKey"

	require := s.Require()

	apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeWorkspace, keyName)
	require.Nil(err)

	res := s.apiKeyService.IsWorkspaceApiKey(apiKey)
	require.True(res)
}

func (s *ApiKeyServiceTestSuite) TestIsWorkspaceApiKey_False() {
	keyName := "clientKey"

	require := s.Require()

	apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeClient, keyName)
	require.Nil(err)

	res := s.apiKeyService.IsWorkspaceApiKey(apiKey)
	require.False(res)
}

func (s *ApiKeyServiceTestSuite) TestIsIsTargetApiKey_True() {
	keyName := "targetKey"

	require := s.Require()

	apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeTarget, keyName)
	require.Nil(err)

	res := s.apiKeyService.IsTargetApiKey(apiKey)
	require.True(res)
}

func (s *ApiKeyServiceTestSuite) TestIsTargetApiKey_False() {
	keyName := "clientKey"

	require := s.Require()

	apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeClient, keyName)
	require.Nil(err)

	res := s.apiKeyService.IsTargetApiKey(apiKey)
	require.False(res)
}
