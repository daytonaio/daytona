// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys_test

import "github.com/daytonaio/daytona/pkg/apikey"

func (s *ApiKeyServiceTestSuite) TestIsValidKey_True() {
	keyName := "api-key"

	require := s.Require()

	apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeProject, keyName)
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

func (s *ApiKeyServiceTestSuite) TestIsProjectApiKey_True() {
	keyName := "projectKey"

	require := s.Require()

	apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeProject, keyName)
	require.Nil(err)

	res := s.apiKeyService.IsProjectApiKey(apiKey)
	require.True(res)
}

func (s *ApiKeyServiceTestSuite) TestIsProjectApiKey_False() {
	keyName := "clientKey"

	require := s.Require()

	apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeClient, keyName)
	require.Nil(err)

	res := s.apiKeyService.IsProjectApiKey(apiKey)
	require.False(res)
}
