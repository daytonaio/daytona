//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type InMemoryWorkspaceStore struct {
	workspaces map[string]*models.Workspace
}

func NewInMemoryWorkspaceStore() stores.WorkspaceStore {
	return &InMemoryWorkspaceStore{
		workspaces: make(map[string]*models.Workspace),
	}
}

func (s *InMemoryWorkspaceStore) List() ([]*models.Workspace, error) {
	workspaces := []*models.Workspace{}
	for _, w := range s.workspaces {
		workspaces = append(workspaces, w)
	}

	return workspaces, nil
}

func (s *InMemoryWorkspaceStore) Find(idOrName string) (*models.Workspace, error) {
	w, ok := s.workspaces[idOrName]
	if !ok {
		for _, w := range s.workspaces {
			if w.Name == idOrName {
				return w, nil
			}
		}
		return nil, stores.ErrWorkspaceNotFound
	}

	return w, nil
}

func (s *InMemoryWorkspaceStore) Save(workspace *models.Workspace) error {
	s.workspaces[workspace.Id] = workspace
	return nil
}

func (s *InMemoryWorkspaceStore) Delete(workspace *models.Workspace) error {
	delete(s.workspaces, workspace.Id)
	return nil
}
