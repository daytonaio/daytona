// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfigs_test

import (
	"testing"

	t_targetconfigs "github.com/daytonaio/daytona/internal/testing/provider/targetconfigs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/targetconfigs"
	"github.com/stretchr/testify/suite"
)

var targetConfig1 *provider.TargetConfig = &provider.TargetConfig{
	Name: "target1",
	ProviderInfo: provider.ProviderInfo{
		Name:    "provider1",
		Version: "v1",
	},
	Options: "",
}

var targetConfig2 *provider.TargetConfig = &provider.TargetConfig{
	Name: "target2",
	ProviderInfo: provider.ProviderInfo{
		Name:    "provider2",
		Version: "v1",
	},
	Options: "",
}

var targetConfig3 *provider.TargetConfig = &provider.TargetConfig{
	Name: "target3",
	ProviderInfo: provider.ProviderInfo{
		Name:    "provider1",
		Version: "v1",
	},
	Options: "",
}

var targetConfig4 *provider.TargetConfig = &provider.TargetConfig{
	Name: "new-target",
	ProviderInfo: provider.ProviderInfo{
		Name:    "provider2",
		Version: "v1",
	},
	Options: "",
}

var expectedConfigs []*provider.TargetConfig
var expectedConfigMap map[string]*provider.TargetConfig

type ProviderTargetServiceTestSuite struct {
	suite.Suite
	targetConfigService targetconfigs.ITargetConfigService
	targetConfigStore   provider.TargetConfigStore
}

func NewProviderTargetServiceTestSuite() *ProviderTargetServiceTestSuite {
	return &ProviderTargetServiceTestSuite{}
}

func (s *ProviderTargetServiceTestSuite) SetupTest() {
	expectedConfigs = []*provider.TargetConfig{
		targetConfig1, targetConfig2, targetConfig3,
	}

	expectedConfigMap = map[string]*provider.TargetConfig{
		targetConfig1.Name: targetConfig1,
		targetConfig2.Name: targetConfig2,
		targetConfig3.Name: targetConfig3,
	}

	s.targetConfigStore = t_targetconfigs.NewInMemoryTargetConfigStore()
	s.targetConfigService = targetconfigs.NewTargetConfigService(targetconfigs.TargetConfigServiceConfig{
		TargetConfigStore: s.targetConfigStore,
	})

	for _, target := range expectedConfigs {
		_ = s.targetConfigService.Save(target)
	}
}

func TestProviderTargetService(t *testing.T) {
	suite.Run(t, NewProviderTargetServiceTestSuite())
}

func (s *ProviderTargetServiceTestSuite) TestList() {
	require := s.Require()

	targetConfigs, err := s.targetConfigService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedConfigs, targetConfigs)
}

func (s *ProviderTargetServiceTestSuite) TestMap() {
	require := s.Require()

	targetConfigsMap, err := s.targetConfigService.Map()
	require.Nil(err)
	require.Equal(expectedConfigMap, targetConfigsMap)
}

func (s *ProviderTargetServiceTestSuite) TestFind() {
	require := s.Require()

	targetConfig, err := s.targetConfigService.Find(&provider.TargetConfigFilter{
		Name: &targetConfig1.Name,
	})
	require.Nil(err)
	require.Equal(targetConfig1, targetConfig)
}

func (s *ProviderTargetServiceTestSuite) TestSetDefault() {
	require := s.Require()

	err := s.targetConfigService.SetDefault(targetConfig2)
	require.Nil(err)

	targetConfig, err := s.targetConfigService.Find(&provider.TargetConfigFilter{
		Name: &targetConfig2.Name,
	})
	require.Nil(err)

	require.Equal(targetConfig2, targetConfig)
}

func (s *ProviderTargetServiceTestSuite) TestSave() {
	expectedConfigs = append(expectedConfigs, targetConfig4)

	require := s.Require()

	err := s.targetConfigService.Save(targetConfig4)
	require.Nil(err)

	targetConfigs, err := s.targetConfigService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedConfigs, targetConfigs)
}

func (s *ProviderTargetServiceTestSuite) TestDelete() {
	expectedConfigs = expectedConfigs[:2]

	require := s.Require()

	err := s.targetConfigService.Delete(targetConfig3)
	require.Nil(err)

	targetConfigs, err := s.targetConfigService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedConfigs, targetConfigs)
}
