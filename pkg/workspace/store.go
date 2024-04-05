// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

type Store interface {
	List() ([]*Workspace, error)
	Find(idOrName string) (*Workspace, error)
	Save(workspace *Workspace) error
	Delete(workspace *Workspace) error
}
