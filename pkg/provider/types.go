// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/workspace"
)

type ProviderInfo struct {
	Name    string  `json:"name" validate:"required"`
	Label   *string `json:"label" validate:"optional"`
	Version string  `json:"version" validate:"required"`
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

type TargetRequest struct {
	TargetConfigOptions string
	Target              *target.Target
}

type WorkspaceRequest struct {
	TargetConfigOptions      string
	ContainerRegistry        *containerregistry.ContainerRegistry
	Workspace                *workspace.Workspace
	GitProviderConfig        *gitprovider.GitProviderConfig
	BuilderImage             string
	BuilderContainerRegistry *containerregistry.ContainerRegistry
}

type TargetConfig struct {
	Name         string       `json:"name" validate:"required"`
	ProviderInfo ProviderInfo `json:"providerInfo" validate:"required"`
	// JSON encoded map of options
	Options   string `json:"options" validate:"required"`
	IsDefault bool   `json:"isDefault" validate:"required"`
} // @name TargetConfig

type TargetConfigManifest map[string]TargetConfigProperty // @name TargetConfigManifest

type TargetConfigPropertyType string

const (
	TargetConfigPropertyTypeString   TargetConfigPropertyType = "string"
	TargetConfigPropertyTypeOption   TargetConfigPropertyType = "option"
	TargetConfigPropertyTypeBoolean  TargetConfigPropertyType = "boolean"
	TargetConfigPropertyTypeInt      TargetConfigPropertyType = "int"
	TargetConfigPropertyTypeFloat    TargetConfigPropertyType = "float"
	TargetConfigPropertyTypeFilePath TargetConfigPropertyType = "file-path"
)

type TargetConfigProperty struct {
	Type        TargetConfigPropertyType
	InputMasked bool
	// A regex string matched with the name of the target config to determine if the property should be disabled
	// If the regex matches the target config name, the property will be disabled
	// E.g. "^local$" will disable the property for the local target
	DisabledPredicate string
	// DefaultValue is converted into the appropriate type based on the Type
	// If the property is a FilePath, the DefaultValue is a path to a directory
	DefaultValue string
	// Brief description of the property
	Description string
	// Options is only used if the Type is TargetConfigPropertyTypeOption
	Options []string
	// Suggestions is an optional list of auto-complete values to assist the user while filling the field
	Suggestions []string
} // @name TargetConfigProperty

type RequirementStatus struct {
	Name   string
	Met    bool
	Reason string
} // @name RequirementStatus
