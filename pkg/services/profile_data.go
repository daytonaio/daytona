// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import "github.com/daytonaio/daytona/pkg/models"

type IProfileDataService interface {
	Get(id string) (*models.ProfileData, error)
	Save(profileData *models.ProfileData) error
	Delete(id string) error
}
