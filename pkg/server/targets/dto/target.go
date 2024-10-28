// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/workspace"
	"github.com/daytonaio/daytona/pkg/target/workspace/buildconfig"
)

type TargetDTO struct {
	target.Target
	Info *target.TargetInfo `json:"info" validate:"optional"`
} //	@name	TargetDTO

type WorkspaceDTO struct {
	workspace.Workspace
	Info *workspace.WorkspaceInfo `json:"info" validate:"optional"`
} //	@name	WorkspaceDTO

type CreateTargetDTO struct {
	Id           string               `json:"id" validate:"required"`
	Name         string               `json:"name" validate:"required"`
	TargetConfig string               `json:"targetConfig" validate:"required"`
	Workspaces   []CreateWorkspaceDTO `json:"workspaces" validate:"required,gt=0,dive"`
} //	@name	CreateTargetDTO

type CreateWorkspaceDTO struct {
	Name                string                   `json:"name" validate:"required"`
	Image               *string                  `json:"image,omitempty" validate:"optional"`
	User                *string                  `json:"user,omitempty" validate:"optional"`
	BuildConfig         *buildconfig.BuildConfig `json:"buildConfig,omitempty" validate:"optional"`
	Source              CreateWorkspaceSourceDTO `json:"source" validate:"required"`
	EnvVars             map[string]string        `json:"envVars" validate:"required"`
	GitProviderConfigId *string                  `json:"gitProviderConfigId" validate:"optional"`
} //	@name	CreateWorkspaceDTO

type CreateWorkspaceSourceDTO struct {
	Repository *gitprovider.GitRepository `json:"repository" validate:"required"`
} // @name CreateWorkspaceSourceDTO
