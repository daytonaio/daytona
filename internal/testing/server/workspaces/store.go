//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import "github.com/daytonaio/daytona/pkg/workspace"

type InMemoryWorkspaceStore struct {
	workspaces map[string]*workspace.Workspace
}

func NewInMemoryWorkspaceStore() workspace.Store {
	return &InMemoryWorkspaceStore{
		workspaces: make(map[string]*workspace.Workspace),
	}
}

func (s *InMemoryWorkspaceStore) List() ([]*workspace.WorkspaceViewDTO, error) {
	workspaceViewDTOs := []*workspace.WorkspaceViewDTO{}
	for _, t := range s.workspaces {
		workspaceViewDTOs = append(workspaceViewDTOs, &workspace.WorkspaceViewDTO{Workspace: *t})
	}

	return workspaceViewDTOs, nil
}

func (s *InMemoryWorkspaceStore) Find(idOrName string) (*workspace.WorkspaceViewDTO, error) {
	t, ok := s.workspaces[idOrName]
	if !ok {
		for _, w := range s.workspaces {
			if w.Name == idOrName {
				return &workspace.WorkspaceViewDTO{Workspace: *w}, nil
			}
		}
		return nil, workspace.ErrWorkspaceNotFound
	}

	return &workspace.WorkspaceViewDTO{Workspace: *t}, nil
}

func (s *InMemoryWorkspaceStore) Save(workspace *workspace.Workspace) error {
	s.workspaces[workspace.Id] = workspace
	return nil
}

func (s *InMemoryWorkspaceStore) Delete(workspace *workspace.Workspace) error {
	delete(s.workspaces, workspace.Id)
	return nil
}
