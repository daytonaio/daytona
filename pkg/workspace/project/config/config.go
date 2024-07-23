// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace/project/build"
)

type ProjectConfig struct {
	Name       string                     `json:"name"`
	Image      string                     `json:"image"`
	User       string                     `json:"user"`
	Build      *build.ProjectBuildConfig  `json:"build"`
	Repository *gitprovider.GitRepository `json:"repository"`
	EnvVars    map[string]string          `json:"-"`
	IsDefault  bool                       `json:"default"`
} // @name ProjectConfig
