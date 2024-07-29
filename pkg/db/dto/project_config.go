// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import "github.com/daytonaio/daytona/pkg/workspace/project/config"

type CreateProjectConfigDTO struct {
	Name       string           `gorm:"primaryKey"`
	Image      string           `json:"image"`
	User       string           `json:"user"`
	Build      *ProjectBuildDTO `json:"build,omitempty" gorm:"serializer:json"`
	Repository RepositoryDTO    `gorm:"serializer:json"`
	EnvVars    string           `json:"envVars"`
	IsDefault  bool             `json:"isDefault"`
}

func ToCreateProjectConfigDTO(projectConfig *config.ProjectConfig) CreateProjectConfigDTO {
	return CreateProjectConfigDTO{
		Name:       projectConfig.Name,
		Image:      projectConfig.Image,
		User:       projectConfig.User,
		Build:      ToProjectBuildDTO(projectConfig.BuildConfig),
		Repository: ToRepositoryDTO(projectConfig.Repository),
		EnvVars:    ToEnvVarsString(projectConfig.EnvVars),
		IsDefault:  projectConfig.IsDefault,
	}
}

func ToProjectConfig(projectConfigDTO CreateProjectConfigDTO) *config.ProjectConfig {
	return &config.ProjectConfig{
		Name:        projectConfigDTO.Name,
		Image:       projectConfigDTO.Image,
		User:        projectConfigDTO.User,
		BuildConfig: ToProjectBuild(projectConfigDTO.Build),
		Repository:  ToRepository(projectConfigDTO.Repository),
		EnvVars:     ToEnvVarsMap(projectConfigDTO.EnvVars),
		IsDefault:   projectConfigDTO.IsDefault,
	}
}
