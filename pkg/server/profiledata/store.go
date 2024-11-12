// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profiledata

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type ProfileDataStore interface {
	Get(id string) (*models.ProfileData, error)
	Save(profileData *models.ProfileData) error
	Delete(id string) error
}

const ProfileDataId = "profile_data"

var (
	ErrProfileDataNotFound = errors.New("profile data not found")
)

func IsProfileDataNotFound(err error) bool {
	return err.Error() == ErrProfileDataNotFound.Error()
}
