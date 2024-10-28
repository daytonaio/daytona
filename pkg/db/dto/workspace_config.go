// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/target/workspace/config"
)

type WorkspaceConfigDTO struct {
	Name                string             `gorm:"primaryKey"`
	Image               string             `json:"image"`
	User                string             `json:"user"`
	Build               *WorkspaceBuildDTO `json:"build,omitempty" gorm:"serializer:json"`
	RepositoryUrl       string             `json:"repositoryUrl"`
	EnvVars             map[string]string  `json:"envVars" gorm:"serializer:json"`
	Prebuilds           []PrebuildDTO      `gorm:"serializer:json"`
	IsDefault           bool               `json:"isDefault"`
	GitProviderConfigId *string            `json:"gitProviderConfigId" validate:"optional"`
}

type PrebuildDTO struct {
	Id             string   `json:"id"`
	Branch         string   `json:"branch"`
	CommitInterval *int     `json:"commitInterval,omitempty"`
	TriggerFiles   []string `json:"triggerFiles,omitempty"`
	Retention      int      `json:"retention"`
}

func ToWorkspaceConfigDTO(workspaceConfig *config.WorkspaceConfig) WorkspaceConfigDTO {
	prebuilds := []PrebuildDTO{}
	for _, prebuild := range workspaceConfig.Prebuilds {
		prebuilds = append(prebuilds, ToPrebuildDTO(prebuild))
	}

	return WorkspaceConfigDTO{
		Name:                workspaceConfig.Name,
		Image:               workspaceConfig.Image,
		User:                workspaceConfig.User,
		Build:               ToWorkspaceBuildDTO(workspaceConfig.BuildConfig),
		RepositoryUrl:       workspaceConfig.RepositoryUrl,
		EnvVars:             workspaceConfig.EnvVars,
		Prebuilds:           prebuilds,
		IsDefault:           workspaceConfig.IsDefault,
		GitProviderConfigId: workspaceConfig.GitProviderConfigId,
	}
}

func ToWorkspaceConfig(workspaceConfigDTO WorkspaceConfigDTO) *config.WorkspaceConfig {
	prebuilds := []*config.PrebuildConfig{}
	for _, prebuildDTO := range workspaceConfigDTO.Prebuilds {
		prebuilds = append(prebuilds, ToPrebuild(prebuildDTO))
	}

	return &config.WorkspaceConfig{
		Name:                workspaceConfigDTO.Name,
		Image:               workspaceConfigDTO.Image,
		User:                workspaceConfigDTO.User,
		BuildConfig:         ToWorkspaceBuild(workspaceConfigDTO.Build),
		RepositoryUrl:       workspaceConfigDTO.RepositoryUrl,
		EnvVars:             workspaceConfigDTO.EnvVars,
		Prebuilds:           prebuilds,
		IsDefault:           workspaceConfigDTO.IsDefault,
		GitProviderConfigId: workspaceConfigDTO.GitProviderConfigId,
	}
}

func ToPrebuildDTO(prebuild *config.PrebuildConfig) PrebuildDTO {
	return PrebuildDTO{
		Id:             prebuild.Id,
		Branch:         prebuild.Branch,
		CommitInterval: prebuild.CommitInterval,
		TriggerFiles:   prebuild.TriggerFiles,
		Retention:      prebuild.Retention,
	}
}

func ToPrebuild(prebuildDTO PrebuildDTO) *config.PrebuildConfig {
	return &config.PrebuildConfig{
		Id:             prebuildDTO.Id,
		Branch:         prebuildDTO.Branch,
		CommitInterval: prebuildDTO.CommitInterval,
		TriggerFiles:   prebuildDTO.TriggerFiles,
		Retention:      prebuildDTO.Retention,
	}
}
