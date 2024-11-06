// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfigs_test

import (
	"testing"

	t_targetconfigs "github.com/daytonaio/daytona/internal/testing/provider/targetconfigs"
	"github.com/daytonaio/daytona/pkg/server/targetconfigs"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/config"
	"github.com/stretchr/testify/suite"
)

var targetConfig1 *config.TargetConfig = &config.TargetConfig{
	Name: "targetConfig1",
	ProviderInfo: target.ProviderInfo{
		Name:    "provider1",
		Version: "v1",
	},
	Options: "",
}

var targetConfig2 *config.TargetConfig = &config.TargetConfig{
	Name: "targetConfig2",
	ProviderInfo: target.ProviderInfo{
		Name:    "provider2",
		Version: "v1",
	},
	Options: "",
}

var targetConfig3 *config.TargetConfig = &config.TargetConfig{
	Name: "targetConfig3",
	ProviderInfo: target.ProviderInfo{
		Name:    "provider3",
		Version: "v1",
	},
	Options: "",
}

var targetConfig4 *config.TargetConfig = &config.TargetConfig{
	Name: "newTargetConfig",
	ProviderInfo: target.ProviderInfo{
		Name:    "provider2",
		Version: "v1",
	},
	Options: "",
}

var expectedConfigs []*config.TargetConfig
var expectedConfigMap map[string]*config.TargetConfig

type TargetConfigServiceTestSuite struct {
	suite.Suite
	targetConfigService targetconfigs.ITargetConfigService
	targetConfigStore   config.TargetConfigStore
}

func NewTargetConfigServiceTestSuite() *TargetConfigServiceTestSuite {
	return &TargetConfigServiceTestSuite{}
}

func (s *TargetConfigServiceTestSuite) SetupTest() {
	expectedConfigs = []*config.TargetConfig{
		targetConfig1, targetConfig2, targetConfig3,
	}

	expectedConfigMap = map[string]*config.TargetConfig{
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

	targetConfig, err := s.targetConfigService.Find(&config.TargetConfigFilter{
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
