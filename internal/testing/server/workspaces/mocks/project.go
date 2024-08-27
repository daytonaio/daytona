//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/workspace/project"
)

var MockProject = project.Project{
	ProjectConfig: MockProjectConfig,
}
