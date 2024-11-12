// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

type ApiKeyType string

const (
	ApiKeyTypeClient    ApiKeyType = "client"
	ApiKeyTypeWorkspace ApiKeyType = "workspace"
	ApiKeyTypeTarget    ApiKeyType = "target"
)

type ApiKey struct {
	KeyHash string     `json:"keyHash" validate:"required" gorm:"primaryKey"`
	Type    ApiKeyType `json:"type" validate:"required"`
	// Workspace or client name
	Name string `json:"name" validate:"required" gorm:"uniqueIndex"`
} // @name ApiKey
