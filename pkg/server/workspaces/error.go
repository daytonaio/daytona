// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"errors"
)

var (
	ErrWorkspaceAlreadyExists = errors.New("workspace already exists")
	ErrInvalidWorkspaceName   = errors.New("name is not a valid alphanumeric string")
	ErrWorkspaceNotFound      = errors.New("workspace not found")
	ErrProjectNotFound        = errors.New("project not found")
	ErrInvalidProjectName     = errors.New("project name is not valid. Only [a-zA-Z0-9-_.] are allowed")
	ErrInvalidProjectConfig   = errors.New("project config is invalid")
)

func IsWorkspaceAlreadyExists(err error) bool {
	return err.Error() == ErrWorkspaceAlreadyExists.Error()
}

func IsWorkspaceNotFound(err error) bool {
	return err.Error() == ErrWorkspaceNotFound.Error()
}

func IsProjectNotFound(err error) bool {
	return err.Error() == ErrProjectNotFound.Error()
}

func IsInvalidWorkspaceName(err error) bool {
	return err.Error() == ErrInvalidWorkspaceName.Error()
}
