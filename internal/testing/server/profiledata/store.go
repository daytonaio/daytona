//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profiledata

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/profiledata"
)

type InMemoryProfileDataStore struct {
	profileData *models.ProfileData
}

func NewInMemoryProfileDataStore() profiledata.ProfileDataStore {
	return &InMemoryProfileDataStore{
		profileData: nil,
	}
}

func (s *InMemoryProfileDataStore) Get(id string) (*models.ProfileData, error) {
	if s.profileData == nil {
		return nil, profiledata.ErrProfileDataNotFound
	}

	return s.profileData, nil
}

func (s *InMemoryProfileDataStore) Save(profileData *models.ProfileData) error {
	s.profileData = profileData
	return nil
}

func (s *InMemoryProfileDataStore) Delete(id string) error {
	s.profileData = nil
	return nil
}
