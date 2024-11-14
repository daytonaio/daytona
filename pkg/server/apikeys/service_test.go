// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys_test

import (
	"testing"

	t_apikeys "github.com/daytonaio/daytona/internal/testing/server/apikeys"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/stretchr/testify/suite"
)

var clientKeyNames []string = []string{"client1", "client2", "client3"}
var workspaceKeyNames []string = []string{"workspace1", "workspace2"}

type ApiKeyServiceTestSuite struct {
	suite.Suite
	apiKeyService services.IApiKeyService
	apiKeyStore   stores.ApiKeyStore
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
		_, _ = s.apiKeyService.Generate(models.ApiKeyTypeClient, keyName)
	}

	for _, keyName := range workspaceKeyNames {
		_, _ = s.apiKeyService.Generate(models.ApiKeyTypeWorkspace, keyName)
	}
}

func TestApiKeyService(t *testing.T) {
	suite.Run(t, NewApiKeyServiceTestSuite())
}
