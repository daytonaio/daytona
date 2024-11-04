// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import "errors"

type Store interface {
	List() ([]*WorkspaceViewDTO, error)
	Find(idOrName string) (*WorkspaceViewDTO, error)
	Save(workspace *Workspace) error
	Delete(workspace *Workspace) error
}

type WorkspaceViewDTO struct {
	Workspace
	TargetName string `json:"targetName" validate:"required"`
} // @name WorkspaceViewDTO

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
)

func IsWorkspaceNotFound(err error) bool {
	return err.Error() == ErrWorkspaceNotFound.Error()
}
