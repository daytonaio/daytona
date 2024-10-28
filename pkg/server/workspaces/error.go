// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"errors"
)

var (
	ErrWorkspaceAlreadyExists = errors.New("workspace already exists")
	ErrWorkspaceNotFound      = errors.New("workspace not found")
	ErrInvalidWorkspaceName   = errors.New("workspace name is not valid. Only [a-zA-Z0-9-_.] are allowed")
	ErrInvalidWorkspaceConfig = errors.New("workspace config is invalid")
)
