// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profiledata_test

import (
	"testing"

	t_profiledata "github.com/daytonaio/daytona/internal/testing/server/profiledata"
	. "github.com/daytonaio/daytona/pkg/profiledata"
	"github.com/daytonaio/daytona/pkg/server/profiledata"
	"github.com/stretchr/testify/suite"
)

type ProfileDataServiceTestSuite struct {
	suite.Suite
	profileDataService profiledata.IProfileDataService
	profileDataStore   Store
}

func NewApiKeyServiceTestSuite() *ProfileDataServiceTestSuite {
	return &ProfileDataServiceTestSuite{}
}

func (s *ProfileDataServiceTestSuite) SetupTest() {
	s.profileDataStore = t_profiledata.NewInMemoryProfileDataStore()
	s.profileDataService = profiledata.NewProfileDataService(profiledata.ProfileDataServiceConfig{
		ProfileDataStore: s.profileDataStore,
	})
}

func TestApiKeyService(t *testing.T) {
	suite.Run(t, NewApiKeyServiceTestSuite())
}

func (s *ProfileDataServiceTestSuite) TestReturnsProfileDataNotFound() {
	profileData, err := s.profileDataService.Get()
	s.Require().Nil(profileData)
	s.Require().True(IsProfileDataNotFound(err))
}

func (s *ProfileDataServiceTestSuite) TestSaveProfileData() {
	profileData := &ProfileData{
		EnvVars: map[string]string{
			"key1": "value1",
		},
	}

	err := s.profileDataService.Save(profileData)
	s.Require().Nil(err)

	profileDataFromStore, err := s.profileDataStore.Get()
	s.Require().Nil(err)
	s.Require().NotNil(profileDataFromStore)
	s.Require().Equal(profileData, profileDataFromStore)
}

func (s *ProfileDataServiceTestSuite) TestDeleteProfileData() {
	profileData := &ProfileData{
		EnvVars: map[string]string{
			"key1": "value1",
		},
	}

	err := s.profileDataService.Save(profileData)
	s.Require().Nil(err)

	err = s.profileDataService.Delete()
	s.Require().Nil(err)

	profileDataFromStore, err := s.profileDataStore.Get()
	s.Require().Nil(profileDataFromStore)
	s.Require().True(IsProfileDataNotFound(err))
}
