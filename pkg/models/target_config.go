// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

type TargetConfig struct {
	Name         string       `json:"name" validate:"required" gorm:"primaryKey"`
	ProviderInfo ProviderInfo `json:"providerInfo" validate:"required" gorm:"serializer:json"`
	// JSON encoded map of options
	Options string `json:"options" validate:"required"`
} // @name TargetConfig
