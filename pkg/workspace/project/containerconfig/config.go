// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerconfig

type ContainerConfig struct {
	Image string `json:"image" validate:"required"`
	User  string `json:"user" validate:"required"`
} // @name ContainerConfig
