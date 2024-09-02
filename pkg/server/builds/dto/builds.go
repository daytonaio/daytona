// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace/project/buildconfig"
)

type BuildCreationData struct {
	Image       string                     `json:"image" validate:"required"`
	User        string                     `json:"user" validate:"required"`
	BuildConfig *buildconfig.BuildConfig   `json:"buildConfig" validate:"optional"`
	Repository  *gitprovider.GitRepository `json:"repository" validate:"optional"`
	EnvVars     map[string]string          `json:"envVars" validate:"required"`
	PrebuildId  string                     `json:"prebuildId" validate:"required"`
} // @name BuildCreationData
