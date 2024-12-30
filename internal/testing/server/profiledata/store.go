//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profiledata

import (
	"github.com/daytonaio/daytona/pkg/profiledata"
)

type InMemoryProfileDataStore struct {
	profileData *profiledata.ProfileData
}

func NewInMemoryProfileDataStore() profiledata.Store {
	return &InMemoryProfileDataStore{
		profileData: nil,
	}
}

func (s *InMemoryProfileDataStore) Get() (*profiledata.ProfileData, error) {
	if s.profileData == nil {
		return nil, profiledata.ErrProfileDataNotFound
	}

	return s.profileData, nil
}

func (s *InMemoryProfileDataStore) Save(profileData *profiledata.ProfileData) error {
	s.profileData = profileData
	return nil
}

func (s *InMemoryProfileDataStore) Delete() error {
	s.profileData = nil
	return nil
}
