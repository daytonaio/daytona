// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

type ProviderInfo struct {
	RunnerId             string               `json:"runnerId" validate:"required"`
	RunnerName           string               `json:"runnerName" validate:"required"`
	Name                 string               `json:"name" validate:"required"`
	Version              string               `json:"version" validate:"required"`
	AgentlessTarget      bool                 `json:"agentlessTarget" validate:"optional"`
	Label                *string              `json:"label" validate:"optional"`
	TargetConfigManifest TargetConfigManifest `json:"targetConfigManifest" validate:"required"`
} // @name ProviderInfo

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
