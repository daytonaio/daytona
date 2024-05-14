// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profiledata

import (
	. "github.com/daytonaio/daytona/pkg/profiledata"
)

type IProfileDataService interface {
	Get() (*ProfileData, error)
	Save(profileData *ProfileData) error
	Delete() error
}

type ProfileDataServiceConfig struct {
	ProfileDataStore Store
}

func NewProfileDataService(config ProfileDataServiceConfig) IProfileDataService {
	return &ProfileDataService{
		profileDataStore: config.ProfileDataStore,
	}
}

type ProfileDataService struct {
	profileDataStore Store
}

func (s *ProfileDataService) Get() (*ProfileData, error) {
	return s.profileDataStore.Get()
}

func (s *ProfileDataService) Save(profileData *ProfileData) error {
	return s.profileDataStore.Save(profileData)
}

func (s *ProfileDataService) Delete() error {
	return s.profileDataStore.Delete()
}
