// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfigs_test

import (
	"testing"

	t_targetconfigs "github.com/daytonaio/daytona/internal/testing/server/targetconfigs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/targetconfigs"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/stretchr/testify/suite"
)

var targetConfig1 *models.TargetConfig = &models.TargetConfig{
	Name: "targetConfig1",
	ProviderInfo: models.ProviderInfo{
		Name:    "provider1",
		Version: "v1",
	},
	Options: "",
}

var targetConfig2 *models.TargetConfig = &models.TargetConfig{
	Name: "targetConfig2",
	ProviderInfo: models.ProviderInfo{
		Name:    "provider2",
		Version: "v1",
	},
	Options: "",
}

var targetConfig3 *models.TargetConfig = &models.TargetConfig{
	Name: "targetConfig3",
	ProviderInfo: models.ProviderInfo{
		Name:    "provider3",
		Version: "v1",
	},
	Options: "",
}

var targetConfig4 *models.TargetConfig = &models.TargetConfig{
	Name: "newTargetConfig",
	ProviderInfo: models.ProviderInfo{
		Name:    "provider2",
		Version: "v1",
	},
	Options: "",
}

var expectedConfigs []*models.TargetConfig
var expectedConfigMap map[string]*models.TargetConfig

type TargetConfigServiceTestSuite struct {
	suite.Suite
	targetConfigService services.ITargetConfigService
	targetConfigStore   stores.TargetConfigStore
}

func NewTargetConfigServiceTestSuite() *TargetConfigServiceTestSuite {
	return &TargetConfigServiceTestSuite{}
}

func (s *TargetConfigServiceTestSuite) SetupTest() {
	expectedConfigs = []*models.TargetConfig{
		targetConfig1, targetConfig2, targetConfig3,
	}

	expectedConfigMap = map[string]*models.TargetConfig{
		targetConfig1.Name: targetConfig1,
		targetConfig2.Name: targetConfig2,
		targetConfig3.Name: targetConfig3,
	}

	s.targetConfigStore = t_targetconfigs.NewInMemoryTargetConfigStore()
	s.targetConfigService = targetconfigs.NewTargetConfigService(targetconfigs.TargetConfigServiceConfig{
		TargetConfigStore: s.targetConfigStore,
	})

	for _, targetConfig := range expectedConfigs {
		_ = s.targetConfigService.Save(targetConfig)
	}
}

func TestTargetConfigService(t *testing.T) {
	suite.Run(t, NewTargetConfigServiceTestSuite())
}

func (s *TargetConfigServiceTestSuite) TestList() {
	require := s.Require()

	targetConfigs, err := s.targetConfigService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedConfigs, targetConfigs)
}

func (s *TargetConfigServiceTestSuite) TestMap() {
	require := s.Require()

	targetConfigsMap, err := s.targetConfigService.Map()
	require.Nil(err)
	require.Equal(expectedConfigMap, targetConfigsMap)
}

func (s *TargetConfigServiceTestSuite) TestFind() {
	require := s.Require()

	targetConfig, err := s.targetConfigService.Find(&stores.TargetConfigFilter{
		Name: &targetConfig1.Name,
	})
	require.Nil(err)
	require.Equal(targetConfig1, targetConfig)
}

func (s *TargetConfigServiceTestSuite) TestSave() {
	expectedConfigs = append(expectedConfigs, targetConfig4)

	require := s.Require()

	err := s.targetConfigService.Save(targetConfig4)
	require.Nil(err)

	targetConfigs, err := s.targetConfigService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedConfigs, targetConfigs)
}

func (s *TargetConfigServiceTestSuite) TestDelete() {
	expectedConfigs = expectedConfigs[:2]

	require := s.Require()

	err := s.targetConfigService.Delete(targetConfig3)
	require.Nil(err)

	targetConfigs, err := s.targetConfigService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedConfigs, targetConfigs)
}
