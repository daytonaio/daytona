// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

type TargetConfig struct {
	Id           string       `json:"id" validate:"required" gorm:"primaryKey"`
	Name         string       `json:"name" validate:"required"`
	ProviderInfo ProviderInfo `json:"providerInfo" validate:"required" gorm:"serializer:json"`
	// JSON encoded map of options
	Options string `json:"options" validate:"required"`
	Deleted bool   `json:"deleted" validate:"required"`
} // @name TargetConfig
