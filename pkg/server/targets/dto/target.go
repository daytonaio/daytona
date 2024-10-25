// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/project"
	"github.com/daytonaio/daytona/pkg/target/project/buildconfig"
)

type TargetDTO struct {
	target.Target
	Info *target.TargetInfo `json:"info" validate:"optional"`
} //	@name	TargetDTO

type ProjectDTO struct {
	project.Project
	Info *project.ProjectInfo `json:"info" validate:"optional"`
} //	@name	ProjectDTO

type CreateTargetDTO struct {
	Id           string             `json:"id" validate:"required"`
	Name         string             `json:"name" validate:"required"`
	TargetConfig string             `json:"targetConfig" validate:"required"`
	Projects     []CreateProjectDTO `json:"projects" validate:"required,gt=0,dive"`
} //	@name	CreateTargetDTO

type CreateProjectDTO struct {
	Name                string                   `json:"name" validate:"required"`
	Image               *string                  `json:"image,omitempty" validate:"optional"`
	User                *string                  `json:"user,omitempty" validate:"optional"`
	BuildConfig         *buildconfig.BuildConfig `json:"buildConfig,omitempty" validate:"optional"`
	Source              CreateProjectSourceDTO   `json:"source" validate:"required"`
	EnvVars             map[string]string        `json:"envVars" validate:"required"`
	GitProviderConfigId *string                  `json:"gitProviderConfigId" validate:"optional"`
} //	@name	CreateProjectDTO

type CreateProjectSourceDTO struct {
	Repository *gitprovider.GitRepository `json:"repository" validate:"required"`
} // @name CreateProjectSourceDTO
