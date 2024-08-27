// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import "github.com/daytonaio/daytona/pkg/build"

type BuildDTO struct {
	Id      string     `json:"id"`
	Hash    string     `gorm:"primaryKey"`
	State   string     `json:"state"`
	Project ProjectDTO `json:"project" gorm:"serializer:json"`
	User    string     `json:"user"`
	Image   string     `json:"image"`
}

func ToBuildDTO(build *build.Build) BuildDTO {
	return BuildDTO{
		Id:      build.Id,
		Hash:    build.Hash,
		State:   string(build.State),
		Project: ToProjectDTO(&build.Project),
		User:    build.User,
		Image:   build.Image,
	}
}

func ToBuild(buildDTO BuildDTO) *build.Build {
	return &build.Build{
		Id:      buildDTO.Id,
		Hash:    buildDTO.Hash,
		State:   build.BuildState(buildDTO.State),
		Project: *ToProject(buildDTO.Project),
		User:    buildDTO.User,
		Image:   buildDTO.Image,
	}
}
