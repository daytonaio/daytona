// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package providertargets_test

import (
	"testing"

	"github.com/daytonaio/daytona/internal/testing/provider/targets"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/providertargets"
	"github.com/stretchr/testify/suite"
)

var providerTarget1 *provider.ProviderTarget = &provider.ProviderTarget{
	Name: "target1",
	ProviderInfo: provider.ProviderInfo{
		Name:    "provider1",
		Version: "v1",
	},
	Options: "",
}

var providerTarget2 *provider.ProviderTarget = &provider.ProviderTarget{
	Name: "target2",
	ProviderInfo: provider.ProviderInfo{
		Name:    "provider2",
		Version: "v1",
	},
	Options: "",
}

var providerTarget3 *provider.ProviderTarget = &provider.ProviderTarget{
	Name: "target3",
	ProviderInfo: provider.ProviderInfo{
		Name:    "provider1",
		Version: "v1",
	},
	Options: "",
}

var providerTarget4 *provider.ProviderTarget = &provider.ProviderTarget{
	Name: "new-target",
	ProviderInfo: provider.ProviderInfo{
		Name:    "provider2",
		Version: "v1",
	},
	Options: "",
}

var expectedProviderTargets []*provider.ProviderTarget
var expectedProviderTargetsMap map[string]*provider.ProviderTarget

type ProviderTargetServiceTestSuite struct {
	suite.Suite
	providerTargetService providertargets.IProviderTargetService
	targetStore           provider.TargetStore
}

func NewProviderTargetServiceTestSuite() *ProviderTargetServiceTestSuite {
	return &ProviderTargetServiceTestSuite{}
}

func (s *ProviderTargetServiceTestSuite) SetupTest() {
	expectedProviderTargets = []*provider.ProviderTarget{
		providerTarget1, providerTarget2, providerTarget3,
	}

	expectedProviderTargetsMap = map[string]*provider.ProviderTarget{
		providerTarget1.Name: providerTarget1,
		providerTarget2.Name: providerTarget2,
		providerTarget3.Name: providerTarget3,
	}

	s.targetStore = targets.NewInMemoryTargetStore()
	s.providerTargetService = providertargets.NewProviderTargetService(providertargets.ProviderTargetServiceConfig{
		TargetStore: s.targetStore,
	})

	for _, target := range expectedProviderTargets {
		_ = s.providerTargetService.Save(target)
	}
}

func TestProviderTargetService(t *testing.T) {
	suite.Run(t, NewProviderTargetServiceTestSuite())
}

func (s *ProviderTargetServiceTestSuite) TestList() {
	require := s.Require()

	providerTargets, err := s.providerTargetService.List()
	require.Nil(err)
	require.ElementsMatch(expectedProviderTargets, providerTargets)
}

func (s *ProviderTargetServiceTestSuite) TestMap() {
	require := s.Require()

	providerTargetsMap, err := s.providerTargetService.Map()
	require.Nil(err)
	require.Equal(expectedProviderTargetsMap, providerTargetsMap)
}

func (s *ProviderTargetServiceTestSuite) TestFind() {
	require := s.Require()

	providerTarget, err := s.providerTargetService.Find(providerTarget1.Name)
	require.Nil(err)
	require.Equal(providerTarget1, providerTarget)
}

func (s *ProviderTargetServiceTestSuite) TestSave() {
	expectedProviderTargets = append(expectedProviderTargets, providerTarget4)

	require := s.Require()

	err := s.providerTargetService.Save(providerTarget4)
	require.Nil(err)

	providerTargets, err := s.providerTargetService.List()
	require.Nil(err)
	require.ElementsMatch(expectedProviderTargets, providerTargets)
}

func (s *ProviderTargetServiceTestSuite) TestDelete() {
	expectedProviderTargets = expectedProviderTargets[:2]

	require := s.Require()

	err := s.providerTargetService.Delete(providerTarget3)
	require.Nil(err)

	providerTargets, err := s.providerTargetService.List()
	require.Nil(err)
	require.ElementsMatch(expectedProviderTargets, providerTargets)
}
