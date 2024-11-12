// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys_test

import (
	"github.com/daytonaio/daytona/pkg/models"
)

func (s *ApiKeyServiceTestSuite) TestListClientKeys() {
	expectedKeys := []*models.ApiKey{}
	keyNames := []string{}

	keyNames = append(keyNames, clientKeyNames...)
	for _, keyName := range keyNames {
		apiKey, _ := s.apiKeyStore.FindByName(keyName)
		expectedKeys = append(expectedKeys, apiKey)
	}

	require := s.Require()

	keys, err := s.apiKeyService.ListClientKeys()

	require.Nil(err)
	require.ElementsMatch(expectedKeys, keys)
}

func (s *ApiKeyServiceTestSuite) TestRevoke() {
	expectedKeys := []*models.ApiKey{}
	keyNames := []string{}

	keyNames = append(keyNames, clientKeyNames[1:]...)
	keyNames = append(keyNames, workspaceKeyNames...)
	for _, keyName := range keyNames {
		apiKey, _ := s.apiKeyStore.FindByName(keyName)
		expectedKeys = append(expectedKeys, apiKey)
	}

	require := s.Require()

	err := s.apiKeyService.Revoke(clientKeyNames[0])
	require.Nil(err)

	keys, err := s.apiKeyStore.List()
	require.Nil(err)
	require.ElementsMatch(expectedKeys, keys)
}

func (s *ApiKeyServiceTestSuite) TestGenerate() {
	expectedKeys := []*models.ApiKey{}
	keyNames := []string{}

	keyNames = append(keyNames, clientKeyNames...)
	keyNames = append(keyNames, workspaceKeyNames...)
	for _, keyName := range keyNames {
		apiKey, _ := s.apiKeyStore.FindByName(keyName)
		expectedKeys = append(expectedKeys, apiKey)
	}

	keyName := "client"

	require := s.Require()

	_, err := s.apiKeyService.Generate(models.ApiKeyTypeClient, keyName)
	require.Nil(err)

	apiKey, err := s.apiKeyStore.FindByName(keyName)
	require.Nil(err)
	expectedKeys = append(expectedKeys, apiKey)

	apiKeys, err := s.apiKeyStore.List()
	require.Nil(err)
	require.ElementsMatch(expectedKeys, apiKeys)
}
