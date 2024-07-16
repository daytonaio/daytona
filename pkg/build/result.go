// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

type BuildResult struct {
	Hash              string `json:"hash"`
	User              string `json:"user"`
	ImageName         string `json:"imageName"`
	ProjectVolumePath string `json:"projectVolumePath"`
} // @name BuildResult
