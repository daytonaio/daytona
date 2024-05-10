// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import "github.com/daytonaio/daytona/pkg/profiledata"

const ProfileDataId = "profile_data"

type ProfileDataDTO struct {
	Id      string            `gorm:"primaryKey"`
	EnvVars map[string]string `gorm:"serializer:json"`
}

func ToProfileDataDTO(profileData *profiledata.ProfileData) ProfileDataDTO {
	return ProfileDataDTO{
		Id:      ProfileDataId,
		EnvVars: profileData.EnvVars,
	}
}

func ToProfileData(profileDataDTO ProfileDataDTO) *profiledata.ProfileData {
	return &profiledata.ProfileData{
		EnvVars: profileDataDTO.EnvVars,
	}
}
