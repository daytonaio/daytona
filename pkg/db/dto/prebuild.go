// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import "github.com/daytonaio/daytona/pkg/prebuild"

type PrebuildDTO struct {
	Key            string           `gorm:"primaryKey" json:"key"`
	Branch         *string          `json:"branch,omitempty"`
	ProjectConfig  ProjectConfigDTO `gorm:"serializer:json"`
	CommitInterval *int             `json:"commitInterval,omitempty"`
	TriggerFiles   []string         `gorm:"serializer:json"`
}

func ToPrebuildDTO(prebuild *prebuild.Prebuild) PrebuildDTO {
	return PrebuildDTO{
		Key:            prebuild.Key,
		Branch:         &prebuild.Branch,
		ProjectConfig:  ToProjectConfigDTO(&prebuild.ProjectConfig),
		CommitInterval: prebuild.CommitInterval,
		TriggerFiles:   prebuild.TriggerFiles,
	}
}

func ToPrebuild(prebuildDTO PrebuildDTO) *prebuild.Prebuild {
	return &prebuild.Prebuild{
		Key:            prebuildDTO.Key,
		Branch:         *prebuildDTO.Branch,
		ProjectConfig:  *ToProjectConfig(prebuildDTO.ProjectConfig),
		CommitInterval: prebuildDTO.CommitInterval,
		TriggerFiles:   prebuildDTO.TriggerFiles,
	}
}
