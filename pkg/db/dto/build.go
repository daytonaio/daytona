// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"time"

	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/workspace/containerconfig"
)

type BuildDTO struct {
	Id              string                          `json:"id" gorm:"primaryKey"`
	State           string                          `json:"state"`
	Image           *string                         `json:"image,omitempty"`
	User            *string                         `json:"user,omitempty"`
	ContainerConfig containerconfig.ContainerConfig `gorm:"serializer:json"`
	BuildConfig     *WorkspaceBuildDTO              `json:"build,omitempty" gorm:"serializer:json"`
	Repository      RepositoryDTO                   `gorm:"serializer:json"`
	EnvVars         map[string]string               `json:"envVars" gorm:"serializer:json"`
	PrebuildId      string                          `json:"prebuildId"`
	CreatedAt       time.Time                       `json:"createdAt"`
	UpdatedAt       time.Time                       `json:"updatedAt"`
}

func ToBuildDTO(build *build.Build) BuildDTO {
	return BuildDTO{
		Id:              build.Id,
		State:           string(build.State),
		Image:           build.Image,
		User:            build.User,
		ContainerConfig: build.ContainerConfig,
		BuildConfig:     ToWorkspaceBuildDTO(build.BuildConfig),
		Repository:      ToRepositoryDTO(build.Repository),
		EnvVars:         build.EnvVars,
		PrebuildId:      build.PrebuildId,
		CreatedAt:       build.CreatedAt,
		UpdatedAt:       build.UpdatedAt,
	}
}

func ToBuild(buildDTO BuildDTO) *build.Build {
	return &build.Build{
		Id:              buildDTO.Id,
		State:           build.BuildState(buildDTO.State),
		Image:           buildDTO.Image,
		User:            buildDTO.User,
		ContainerConfig: buildDTO.ContainerConfig,
		BuildConfig:     ToWorkspaceBuild(buildDTO.BuildConfig),
		Repository:      ToRepository(buildDTO.Repository),
		EnvVars:         buildDTO.EnvVars,
		PrebuildId:      buildDTO.PrebuildId,
		CreatedAt:       buildDTO.CreatedAt,
		UpdatedAt:       buildDTO.UpdatedAt,
	}
}
