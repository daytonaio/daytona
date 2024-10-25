//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/target/project/buildconfig"
	"github.com/daytonaio/daytona/pkg/target/project/config"
)

var MockProjectConfig = config.ProjectConfig{
	BuildConfig: &buildconfig.BuildConfig{
		Devcontainer: &buildconfig.DevcontainerConfig{
			FilePath: ".devcontainer/devcontainer.json",
		},
	},
	RepositoryUrl: "https://github.com/daytonaio/daytona",
}
