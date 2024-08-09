// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profiledata

type ProfileData struct {
	EnvVars map[string]string `json:"envVars" validate:"required"`
} // @name ProfileData
