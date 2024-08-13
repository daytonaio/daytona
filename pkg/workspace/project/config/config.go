// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace/project/buildconfig"
)

type ProjectConfig struct {
	Name        string                          `json:"name" validate:"required"`
	Image       string                          `json:"image" validate:"required"`
	User        string                          `json:"user" validate:"required"`
	BuildConfig *buildconfig.ProjectBuildConfig `json:"buildConfig,omitempty" validate:"optional"`
	Repository  *gitprovider.GitRepository      `json:"repository" validate:"required"`
	EnvVars     map[string]string               `json:"envVars" validate:"required"`
	IsDefault   bool                            `json:"default" validate:"required"`
} // @name ProjectConfig
