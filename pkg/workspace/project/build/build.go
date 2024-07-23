// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

type DevcontainerConfig struct {
	FilePath string `json:"filePath"`
} // @name DevcontainerConfig

/*
type DockerfileConfig struct {
	Context    string            `json:"context"`
	Dockerfile string            `json:"dockerfile"`
	Args       map[string]string `json:"args"`
} // @name DockerfileConfig
*/

type ProjectBuildConfig struct {
	Devcontainer *DevcontainerConfig `json:"devcontainer"`
	/*
		Dockerfile   *ProjectBuildDockerfile   `json:"dockerfile"`
	*/
} // @name ProjectBuildConfig
