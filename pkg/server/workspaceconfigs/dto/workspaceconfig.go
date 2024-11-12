// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import "github.com/daytonaio/daytona/pkg/models"

type CreateWorkspaceConfigDTO struct {
	Name                string              `json:"name" validate:"required"`
	Image               *string             `json:"image,omitempty" validate:"optional"`
	User                *string             `json:"user,omitempty" validate:"optional"`
	BuildConfig         *models.BuildConfig `json:"buildConfig,omitempty" validate:"optional"`
	RepositoryUrl       string              `json:"repositoryUrl" validate:"required"`
	EnvVars             map[string]string   `json:"envVars" validate:"required"`
	GitProviderConfigId *string             `json:"gitProviderConfigId" validate:"optional"`
} // @name CreateWorkspaceConfigDTO

type PrebuildDTO struct {
	Id                  string   `json:"id" validate:"required"`
	WorkspaceConfigName string   `json:"workspaceConfigName" validate:"required"`
	Branch              string   `json:"branch" validate:"required"`
	CommitInterval      *int     `json:"commitInterval" validate:"optional"`
	TriggerFiles        []string `json:"triggerFiles" validate:"optional"`
	Retention           int      `json:"retention" validate:"required"`
} // @name PrebuildDTO

type CreatePrebuildDTO struct {
	Id             *string  `json:"id" validate:"optional"`
	Branch         string   `json:"branch" validate:"optional"`
	CommitInterval *int     `json:"commitInterval" validate:"optional"`
	TriggerFiles   []string `json:"triggerFiles" validate:"optional"`
	Retention      int      `json:"retention" validate:"required"`
} // @name CreatePrebuildDTO
