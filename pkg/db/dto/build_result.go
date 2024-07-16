// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import "github.com/daytonaio/daytona/pkg/build"

type BuildResultDTO struct {
	Hash              string `gorm:"primaryKey"`
	User              string `json:"user"`
	ImageName         string `json:"imageName"`
	ProjectVolumePath string `json:"projectVolumePath"`
}

func ToBuildResultDTO(buildResult *build.BuildResult) BuildResultDTO {
	return BuildResultDTO{
		Hash:              buildResult.Hash,
		User:              buildResult.User,
		ImageName:         buildResult.ImageName,
		ProjectVolumePath: buildResult.ProjectVolumePath,
	}
}

func ToBuildResult(buildResultDTO BuildResultDTO) *build.BuildResult {
	return &build.BuildResult{
		Hash:              buildResultDTO.Hash,
		User:              buildResultDTO.User,
		ImageName:         buildResultDTO.ImageName,
		ProjectVolumePath: buildResultDTO.ProjectVolumePath,
	}
}
