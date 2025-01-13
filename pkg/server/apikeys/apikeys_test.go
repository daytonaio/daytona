// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys_test

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
)

func (s *ApiKeyServiceTestSuite) TestListClientKeys() {
	expectedKeys := []*services.ApiKeyDTO{}
	keyNames := []string{}

	keyNames = append(keyNames, clientKeyNames...)
	for _, keyName := range keyNames {
		apiKey, _ := s.apiKeyStore.FindByName(context.TODO(), keyName)
		expectedKeys = append(expectedKeys, &services.ApiKeyDTO{
			Type: apiKey.Type,
			Name: apiKey.Name,
		})
	}

	require := s.Require()

	keys, err := s.apiKeyService.ListClientKeys(context.TODO())

	require.Nil(err)
	require.ElementsMatch(expectedKeys, keys)
}

func (s *ApiKeyServiceTestSuite) TestRevoke() {
	expectedKeys := []*models.ApiKey{}
	keyNames := []string{}

	keyNames = append(keyNames, clientKeyNames[1:]...)
	keyNames = append(keyNames, workspaceKeyNames...)
	for _, keyName := range keyNames {
		apiKey, _ := s.apiKeyStore.FindByName(context.TODO(), keyName)
		expectedKeys = append(expectedKeys, apiKey)
	}

	require := s.Require()

	err := s.apiKeyService.Revoke(context.TODO(), clientKeyNames[0])
	require.Nil(err)

	keys, err := s.apiKeyStore.List(context.TODO())
	require.Nil(err)
	require.ElementsMatch(expectedKeys, keys)
}

func (s *ApiKeyServiceTestSuite) TestGenerate() {
	expectedKeys := []*models.ApiKey{}
	keyNames := []string{}

	keyNames = append(keyNames, clientKeyNames...)
	keyNames = append(keyNames, workspaceKeyNames...)
	for _, keyName := range keyNames {
		apiKey, _ := s.apiKeyStore.FindByName(context.TODO(), keyName)
		expectedKeys = append(expectedKeys, apiKey)
	}

	keyName := "client"

	require := s.Require()

	_, err := s.apiKeyService.Generate(context.TODO(), models.ApiKeyTypeClient, keyName)
	require.Nil(err)

	apiKey, err := s.apiKeyStore.FindByName(context.TODO(), keyName)
	require.Nil(err)
	expectedKeys = append(expectedKeys, apiKey)

	apiKeys, err := s.apiKeyStore.List(context.TODO())
	require.Nil(err)
	require.ElementsMatch(expectedKeys, apiKeys)
}
