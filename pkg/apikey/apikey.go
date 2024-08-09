// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

type ApiKeyType string

const (
	ApiKeyTypeClient    ApiKeyType = "client"
	ApiKeyTypeProject   ApiKeyType = "project"
	ApiKeyTypeWorkspace ApiKeyType = "workspace"
)

type ApiKey struct {
	KeyHash string     `json:"keyHash" validate:"required"`
	Type    ApiKeyType `json:"type" validate:"required"`
	// Project or client name
	Name string `json:"name" validate:"required"`
} // @name ApiKey
