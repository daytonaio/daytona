// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type ProviderInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeProviderRequest struct {
	BasePath          string
	ServerDownloadUrl string
	ServerVersion     string
	ServerUrl         string
	NetworkKey        string
	ServerApiUrl      string
	LogsDir           string
}

type WorkspaceRequest struct {
	TargetOptions string
	Workspace     *workspace.Workspace
}

type ProjectRequest struct {
	TargetOptions     string
	ContainerRegistry *containerregistry.ContainerRegistry
	Project           *workspace.Project
}

type ProviderTarget struct {
	Name         string       `json:"name"`
	ProviderInfo ProviderInfo `json:"providerInfo"`
	// JSON encoded map of options
	Options string `json:"options"`
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
}
