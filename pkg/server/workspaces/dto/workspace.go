// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/buildconfig"
)

type WorkspaceDTO struct {
	workspace.WorkspaceViewDTO
	Info *workspace.WorkspaceInfo `json:"info" validate:"optional"`
} //	@name	WorkspaceDTO

type CreateWorkspaceDTO struct {
	Id                  string                   `json:"id" validate:"required"`
	Name                string                   `json:"name" validate:"required"`
	Image               *string                  `json:"image,omitempty" validate:"optional"`
	User                *string                  `json:"user,omitempty" validate:"optional"`
	BuildConfig         *buildconfig.BuildConfig `json:"buildConfig,omitempty" validate:"optional"`
	Source              CreateWorkspaceSourceDTO `json:"source" validate:"required"`
	EnvVars             map[string]string        `json:"envVars" validate:"required"`
	TargetId            string                   `json:"targetId" validate:"required"`
	GitProviderConfigId *string                  `json:"gitProviderConfigId" validate:"optional"`
} //	@name	CreateWorkspaceDTO

type CreateWorkspaceSourceDTO struct {
	Repository *gitprovider.GitRepository `json:"repository" validate:"required"`
} // @name CreateWorkspaceSourceDTO
