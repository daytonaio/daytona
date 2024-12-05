// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

type TargetConfig struct {
	Id           string       `json:"id" validate:"required" gorm:"primaryKey"`
	Name         string       `json:"name" validate:"required" gorm:"not null"`
	ProviderInfo ProviderInfo `json:"providerInfo" validate:"required" gorm:"serializer:json;not null"`
	// JSON encoded map of options
	Options string `json:"options" validate:"required" gorm:"not null"`
	Deleted bool   `json:"deleted" validate:"required" gorm:"not null"`
} // @name TargetConfig
