// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
)

type ProviderInfo struct {
	Name    string `json:"name" validate:"required"`
	Version string `json:"version" validate:"required"`
}

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

type WorkspaceRequest struct {
	TargetOptions string
	Workspace     *workspace.Workspace
}

type ProjectRequest struct {
	TargetOptions     string
	ContainerRegistry *containerregistry.ContainerRegistry
	Project           *project.Project
	GitProviderConfig *gitprovider.GitProviderConfig
}

type ProviderTarget struct {
	Name         string       `json:"name" validate:"required"`
	ProviderInfo ProviderInfo `json:"providerInfo" validate:"required"`
	// JSON encoded map of options
	Options string `json:"options" validate:"required"`
} // @name ProviderTarget

type ProviderTargetManifest map[string]ProviderTargetProperty // @name ProviderTargetManifest

type ProviderTargetPropertyType string

const (
	ProviderTargetPropertyTypeString   ProviderTargetPropertyType = "string"
	ProviderTargetPropertyTypeOption   ProviderTargetPropertyType = "option"
	ProviderTargetPropertyTypeBoolean  ProviderTargetPropertyType = "boolean"
	ProviderTargetPropertyTypeInt      ProviderTargetPropertyType = "int"
	ProviderTargetPropertyTypeFloat    ProviderTargetPropertyType = "float"
	ProviderTargetPropertyTypeFilePath ProviderTargetPropertyType = "file-path"
)

type ProviderTargetProperty struct {
	Type        ProviderTargetPropertyType
	InputMasked bool
	// A regex string matched with the name of the target to determine if the property should be disabled
	// If the regex matches the target name, the property will be disabled
	// E.g. "^local$" will disable the property for the local target
	DisabledPredicate string
	// DefaultValue is converted into the appropriate type based on the Type
	// If the property is a FilePath, the DefaultValue is a path to a directory
	DefaultValue string
	// Brief description of the property
	Description string
	// Options is only used if the Type is ProviderTargetPropertyTypeOption
	Options []string
	// Suggestions is an optional list of auto-complete values to assist the user while filling the field
	Suggestions []string
}

type RequirementStatus struct {
	Name   string
	Met    bool
	Reason string
}
