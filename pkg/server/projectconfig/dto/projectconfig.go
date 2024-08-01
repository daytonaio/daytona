// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace/project/buildconfig"
)

type CreateProjectConfigDTO struct {
	Name        string                          `json:"name"`
	Image       *string                         `json:"image,omitempty"`
	User        *string                         `json:"user,omitempty"`
	BuildConfig *buildconfig.ProjectBuildConfig `json:"buildConfig,omitempty"`
	Source      CreateProjectConfigSourceDTO    `json:"source"`
	EnvVars     map[string]string               `json:"envVars"`
} // @name CreateProjectConfigDTO

type CreateProjectConfigSourceDTO struct {
	Repository *gitprovider.GitRepository `json:"repository"`
} // @name CreateProjectConfigSourceDTO
