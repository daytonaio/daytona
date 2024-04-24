// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys_test

import (
	"testing"

	t_apikeys "github.com/daytonaio/daytona/internal/testing/server/apikeys"
	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/stretchr/testify/suite"
)

var clientKeyNames []string = []string{"client1", "client2", "client3"}
var projectKeyNames []string = []string{"project1", "project2"}

type ApiKeyServiceTestSuite struct {
	suite.Suite
	apiKeyService apikeys.IApiKeyService
	apiKeyStore   apikey.Store
}

func NewApiKeyServiceTestSuite() *ApiKeyServiceTestSuite {
	return &ApiKeyServiceTestSuite{}
}

func (s *ApiKeyServiceTestSuite) SetupTest() {
	s.apiKeyStore = t_apikeys.NewInMemoryApiKeyStore()
	s.apiKeyService = apikeys.NewApiKeyService(apikeys.ApiKeyServiceConfig{
		ApiKeyStore: s.apiKeyStore,
	})

	for _, keyName := range clientKeyNames {
		_, _ = s.apiKeyService.Generate(apikey.ApiKeyTypeClient, keyName)
	}

	for _, keyName := range projectKeyNames {
		_, _ = s.apiKeyService.Generate(apikey.ApiKeyTypeProject, keyName)
	}
}

func TestApiKeyService(t *testing.T) {
	suite.Run(t, NewApiKeyServiceTestSuite())
}
