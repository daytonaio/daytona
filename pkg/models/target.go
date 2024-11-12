// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

type Target struct {
	Id           string       `json:"id" validate:"required" gorm:"primaryKey"`
	Name         string       `json:"name" validate:"required" gorm:"unique"`
	ProviderInfo ProviderInfo `json:"providerInfo" validate:"required" gorm:"serializer:json"`
	// JSON encoded map of options
	Options    string            `json:"options" validate:"required"`
	ApiKey     string            `json:"-"`
	EnvVars    map[string]string `json:"-" gorm:"serializer:json"`
	IsDefault  bool              `json:"default" validate:"required"`
	Workspaces []Workspace       `gorm:"foreignKey:TargetId;references:Id"`
} // @name Target

type TargetInfo struct {
	Name             string `json:"name" validate:"required"`
	ProviderMetadata string `json:"providerMetadata,omitempty" validate:"optional"`
} // @name TargetInfo

type ProviderInfo struct {
	Name    string  `json:"name" validate:"required"`
	Version string  `json:"version" validate:"required"`
	Label   *string `json:"label" validate:"optional"`
} // @name TargetProviderInfo
