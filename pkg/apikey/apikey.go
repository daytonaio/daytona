// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

type ApiKeyType string

const (
	ApiKeyTypeClient  ApiKeyType = "client"
	ApiKeyTypeProject ApiKeyType = "project"
)

type ApiKey struct {
	KeyHash string     `json:"keyHash"`
	Type    ApiKeyType `json:"type"`
	// Project or client name
	Name string `json:"name"`
} // @name ApiKey
