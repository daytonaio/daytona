// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profiledata

import (
	"github.com/daytonaio/daytona/pkg/models"
)

type IProfileDataService interface {
	Get(id string) (*models.ProfileData, error)
	Save(profileData *models.ProfileData) error
	Delete(id string) error
}

type ProfileDataServiceConfig struct {
	ProfileDataStore ProfileDataStore
}

func NewProfileDataService(config ProfileDataServiceConfig) IProfileDataService {
	return &ProfileDataService{
		profileDataStore: config.ProfileDataStore,
	}
}

type ProfileDataService struct {
	profileDataStore ProfileDataStore
}

func (s *ProfileDataService) Get(id string) (*models.ProfileData, error) {
	if id == "" {
		id = ProfileDataId
	}

	return s.profileDataStore.Get(id)
}

func (s *ProfileDataService) Save(profileData *models.ProfileData) error {
	if profileData.Id == "" {
		profileData.Id = ProfileDataId
	}

	return s.profileDataStore.Save(profileData)
}

func (s *ProfileDataService) Delete(id string) error {
	if id == "" {
		id = ProfileDataId
	}

	return s.profileDataStore.Delete(id)
}
