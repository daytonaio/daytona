// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package buildconfig

type DevcontainerConfig struct {
	FilePath string `json:"filePath" validate:"required"`
} // @name DevcontainerConfig

/*
type DockerfileConfig struct {
	Context    string            `json:"context"`
	Dockerfile string            `json:"dockerfile"`
	Args       map[string]string `json:"args"`
} // @name DockerfileConfig
*/

type ProjectBuildConfig struct {
	Devcontainer *DevcontainerConfig `json:"devcontainer,omitempty" validate:"optional"`
	/*
		Dockerfile   *ProjectBuildDockerfile   `json:"dockerfile"`
	*/
} // @name ProjectBuildConfig
