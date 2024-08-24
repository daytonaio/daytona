// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace/project/buildconfig"
)

type CreateProjectConfigDTO struct {
	Name        string                          `json:"name" validate:"required"`
	Image       *string                         `json:"image,omitempty" validate:"optional"`
	User        *string                         `json:"user,omitempty" validate:"optional"`
	BuildConfig *buildconfig.ProjectBuildConfig `json:"buildConfig,omitempty" validate:"optional"`
	Source      CreateProjectConfigSourceDTO    `json:"source" validate:"required"`
	EnvVars     map[string]string               `json:"envVars" validate:"required"`
	Identity    string                          `json:"identity" validate:"required"`
} // @name CreateProjectConfigDTO

type CreateProjectConfigSourceDTO struct {
	Repository *gitprovider.GitRepository `json:"repository" validate:"required"`
} // @name CreateProjectConfigSourceDTO
