//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/models"
)

var MockWorkspaceTemplate = models.WorkspaceTemplate{
	BuildConfig: &models.BuildConfig{
		Devcontainer: &models.DevcontainerConfig{
			FilePath: ".devcontainer/devcontainer.json",
		},
	},
	RepositoryUrl: "https://github.com/daytonaio/daytona",
}
