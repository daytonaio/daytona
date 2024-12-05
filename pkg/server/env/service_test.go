// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env_test

import (
	"context"
	"testing"

	t_envvar "github.com/daytonaio/daytona/internal/testing/server/env"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/env"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/stretchr/testify/suite"
)

type EnvironmentVariableServiceTestSuite struct {
	suite.Suite
	environmentVariableService services.IEnvironmentVariableService
	environmentVariableStore   stores.EnvironmentVariableStore
}

func NewEnvironmentVariableTestSuite() *EnvironmentVariableServiceTestSuite {
	return &EnvironmentVariableServiceTestSuite{}
}

func (s *EnvironmentVariableServiceTestSuite) SetupTest() {
	s.environmentVariableStore = t_envvar.NewInMemoryEnvironmentVariableStore()
	s.environmentVariableService = env.NewEnvironmentVariableService(env.EnvironmentVariableServiceConfig{
		EnvironmentVariableStore: s.environmentVariableStore,
	})
}

func TestEnvironmentVariableService(t *testing.T) {
	suite.Run(t, NewEnvironmentVariableTestSuite())
}

func (s *EnvironmentVariableServiceTestSuite) TestReturnsEnvironmentVariableNotFound() {
	envVar, err := s.environmentVariableService.List(context.TODO())
	s.Require().Nil(envVar)
	s.Require().True(stores.IsEnvironmentVariableNotFound(err))
}

func (s *EnvironmentVariableServiceTestSuite) TestSaveEnvironmentVariable() {
	envVar := &models.EnvironmentVariable{
		Key:   "key1",
		Value: "value1",
	}

	err := s.environmentVariableService.Save(context.TODO(), envVar)
	s.Require().Nil(err)

	envVarsFromStore, err := s.environmentVariableStore.List(context.TODO())
	s.Require().Nil(err)
	s.Require().NotNil(envVarsFromStore)
	s.Require().Equal(envVar, envVarsFromStore)
}

func (s *EnvironmentVariableServiceTestSuite) TestDeleteEnvironmentVariable() {
	envVar := &models.EnvironmentVariable{
		Key:   "key1",
		Value: "value1",
	}

	err := s.environmentVariableService.Save(context.TODO(), envVar)
	s.Require().Nil(err)

	err = s.environmentVariableService.Delete(context.TODO(), envVar.Key)
	s.Require().Nil(err)

	EnvVarsFromStore, err := s.environmentVariableStore.List(context.TODO())
	s.Require().Nil(EnvVarsFromStore)
	s.Require().True(stores.IsEnvironmentVariableNotFound(err))
}
