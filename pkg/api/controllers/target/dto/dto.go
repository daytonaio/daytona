// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import "github.com/daytonaio/daytona/pkg/target/workspace"

type SetWorkspaceState struct {
	Uptime    uint64               `json:"uptime" validate:"required"`
	GitStatus *workspace.GitStatus `json:"gitStatus,omitempty" validate:"optional"`
} // @name SetWorkspaceState
