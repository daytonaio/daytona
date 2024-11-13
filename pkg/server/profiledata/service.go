// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profiledata

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
)

type ProfileDataServiceConfig struct {
	ProfileDataStore stores.ProfileDataStore
}

func NewProfileDataService(config ProfileDataServiceConfig) services.IProfileDataService {
	return &ProfileDataService{
		profileDataStore: config.ProfileDataStore,
	}
}

type ProfileDataService struct {
	profileDataStore stores.ProfileDataStore
}

func (s *ProfileDataService) Get(id string) (*models.ProfileData, error) {
	if id == "" {
		id = stores.ProfileDataId
	}

	return s.profileDataStore.Get(id)
}

func (s *ProfileDataService) Save(profileData *models.ProfileData) error {
	if profileData.Id == "" {
		profileData.Id = stores.ProfileDataId
	}

	return s.profileDataStore.Save(profileData)
}

func (s *ProfileDataService) Delete(id string) error {
	if id == "" {
		id = stores.ProfileDataId
	}

	return s.profileDataStore.Delete(id)
}
