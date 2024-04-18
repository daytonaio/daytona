//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/workspace"
)

type InMemoryWorkspaceStore struct {
	workspaces map[string]*workspace.Workspace
}

func NewInMemoryWorkspaceStore() workspace.Store {
	return &InMemoryWorkspaceStore{
		workspaces: make(map[string]*workspace.Workspace),
	}
}

func (s *InMemoryWorkspaceStore) List() ([]*workspace.Workspace, error) {
	workspaces := []*workspace.Workspace{}
	for _, w := range s.workspaces {
		workspaces = append(workspaces, w)
	}

	return workspaces, nil
}

func (s *InMemoryWorkspaceStore) Find(idOrName string) (*workspace.Workspace, error) {
	workspace, ok := s.workspaces[idOrName]
	if !ok {
		for _, w := range s.workspaces {
			if w.Name == idOrName {
				return w, nil
			}
		}
		return nil, errors.New("workspace not found")
	}

	return workspace, nil
}

func (s *InMemoryWorkspaceStore) Save(workspace *workspace.Workspace) error {
	s.workspaces[workspace.Id] = workspace
	return nil
}

func (s *InMemoryWorkspaceStore) Delete(workspace *workspace.Workspace) error {
	delete(s.workspaces, workspace.Id)
	return nil
}
