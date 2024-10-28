// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package buildconfig

type BuildConfig struct {
	Devcontainer *DevcontainerConfig `json:"devcontainer,omitempty" validate:"optional"`
	CachedBuild  *CachedBuild        `json:"cachedBuild,omitempty" validate:"optional"`
} // @name BuildConfig

type DevcontainerConfig struct {
	FilePath string `json:"filePath" validate:"required"`
} // @name DevcontainerConfig

type CachedBuild struct {
	User  string `json:"user" validate:"required"`
	Image string `json:"image" validate:"required"`
} // @name CachedBuild
