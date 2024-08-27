//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace/project/buildconfig"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
)

var MockProjectConfig = config.ProjectConfig{
	BuildConfig: &buildconfig.ProjectBuildConfig{
		Devcontainer: &buildconfig.DevcontainerConfig{
			FilePath: ".devcontainer/devcontainer.json",
		},
	},
	Repository: &gitprovider.GitRepository{
		Url: "url",
	},
}
