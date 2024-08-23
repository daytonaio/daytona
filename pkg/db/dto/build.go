// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"time"

	"github.com/daytonaio/daytona/pkg/build"
)

type BuildDTO struct {
	Id          string            `json:"id" gorm:"primaryKey"`
	State       string            `json:"state"`
	Image       string            `json:"image"`
	User        string            `json:"user"`
	BuildConfig *ProjectBuildDTO  `json:"build,omitempty" gorm:"serializer:json"`
	Repository  RepositoryDTO     `gorm:"serializer:json"`
	EnvVars     map[string]string `json:"envVars" gorm:"serializer:json"`
	PrebuildId  string            `json:"prebuildId"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

func ToBuildDTO(build *build.Build) BuildDTO {
	return BuildDTO{
		Id:          build.Id,
		State:       string(build.State),
		Image:       build.Image,
		User:        build.User,
		BuildConfig: ToProjectBuildDTO(build.BuildConfig),
		Repository:  ToRepositoryDTO(build.Repository),
		EnvVars:     build.EnvVars,
		PrebuildId:  build.PrebuildId,
		CreatedAt:   build.CreatedAt,
		UpdatedAt:   build.UpdatedAt,
	}
}

func ToBuild(buildDTO BuildDTO) *build.Build {
	return &build.Build{
		Id:          buildDTO.Id,
		State:       build.BuildState(buildDTO.State),
		Image:       buildDTO.Image,
		User:        buildDTO.User,
		BuildConfig: ToProjectBuild(buildDTO.BuildConfig),
		Repository:  ToRepository(buildDTO.Repository),
		EnvVars:     buildDTO.EnvVars,
		PrebuildId:  buildDTO.PrebuildId,
		CreatedAt:   buildDTO.CreatedAt,
		UpdatedAt:   buildDTO.UpdatedAt,
	}
}
