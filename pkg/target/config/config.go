// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import "github.com/daytonaio/daytona/pkg/target"

type TargetConfig struct {
	Name         string              `json:"name" validate:"required"`
	ProviderInfo target.ProviderInfo `json:"providerInfo" validate:"required"`
	// JSON encoded map of options
	Options string `json:"options" validate:"required"`
} // @name TargetConfig
