// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"github.com/daytonaio/daytona/pkg/models"
)

type InitializeProviderRequest struct {
	BasePath           string
	DaytonaDownloadUrl string
	DaytonaVersion     string
	LogsDir            string

	NetworkKey string
	ServerUrl  string
	ApiUrl     string
	// ServerPort is used if the target supports direct server access
	ServerPort uint32
	// ApiPort is used if the target supports direct server access
	ApiPort uint32
}

type TargetRequest struct {
	Target *models.Target
}

type WorkspaceRequest struct {
	BuilderImage             string
	BuilderContainerRegistry *models.ContainerRegistry
	ContainerRegistry        *models.ContainerRegistry
	Workspace                *models.Workspace
	GitProviderConfig        *models.GitProviderConfig
}

type TargetConfig struct {
	Name string `json:"name" validate:"required"`
	// JSON encoded map of options
	Options string `json:"options" validate:"required"`
} // @name TargetConfig

type RequirementStatus struct {
	Name   string
	Met    bool
	Reason string
} // @name RequirementStatus
