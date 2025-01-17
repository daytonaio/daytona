//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"

	"github.com/daytonaio/daytona/internal/testing/common"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type InMemoryWorkspaceStore struct {
	common.InMemoryStore
	jobStore   stores.JobStore
	workspaces map[string]*models.Workspace
}

func NewInMemoryWorkspaceStore(jobStore stores.JobStore) stores.WorkspaceStore {
	return &InMemoryWorkspaceStore{
		workspaces: make(map[string]*models.Workspace),
		jobStore:   jobStore,
	}
}

func (s *InMemoryWorkspaceStore) List(ctx context.Context) ([]*models.Workspace, error) {
	workspaces := []*models.Workspace{}
	jobs, err := s.jobMap(ctx)
	if err != nil {
		return nil, err
	}

	for _, w := range s.workspaces {
		w.LastJob = jobs[w.Id]
		workspaces = append(workspaces, w)
	}

	return workspaces, nil
}

func (s *InMemoryWorkspaceStore) Find(ctx context.Context, idOrName string) (*models.Workspace, error) {
	jobs, err := s.jobMap(ctx)
	if err != nil {
		return nil, err
	}

	w, ok := s.workspaces[idOrName]
	if !ok {
		for _, w := range s.workspaces {
			if w.Name == idOrName {
				w.LastJob = jobs[w.Id]
				return w, nil
			}
		}
		return nil, stores.ErrWorkspaceNotFound
	}
	w.LastJob = jobs[w.Id]

	return w, nil
}

func (s *InMemoryWorkspaceStore) Save(ctx context.Context, workspace *models.Workspace) error {
	s.workspaces[workspace.Id] = workspace
	return nil
}

func (s *InMemoryWorkspaceStore) Delete(ctx context.Context, workspace *models.Workspace) error {
	delete(s.workspaces, workspace.Id)
	return nil
}

func (s *InMemoryWorkspaceStore) jobMap(ctx context.Context) (map[string]*models.Job, error) {
	jobs, err := s.jobStore.List(ctx, &stores.JobFilter{
		ResourceType: util.Pointer(models.ResourceTypeWorkspace),
	})
	if err != nil {
		return nil, err
	}

	jobMap := make(map[string]*models.Job)
	for _, j := range jobs {
		jobMap[j.ResourceId] = j
	}

	return jobMap, nil
}
