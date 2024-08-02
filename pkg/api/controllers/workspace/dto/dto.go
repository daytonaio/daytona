// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/workspace/project"
)

type SetProjectState struct {
	Uptime    uint64            `json:"uptime"`
	GitStatus project.GitStatus `json:"gitStatus"`
} // @name SetProjectState
