//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/testing/common"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type InMemoryTargetStore struct {
	common.InMemoryStore
	targets  map[string]*models.Target
	jobStore stores.JobStore
}

func NewInMemoryTargetStore(jobStore stores.JobStore) stores.TargetStore {
	return &InMemoryTargetStore{
		targets:  make(map[string]*models.Target),
		jobStore: jobStore,
	}
}

func (s *InMemoryTargetStore) List(ctx context.Context, filter *stores.TargetFilter) ([]*models.Target, error) {
	return s.processFilters(filter)
}

func (s *InMemoryTargetStore) Find(ctx context.Context, filter *stores.TargetFilter) (*models.Target, error) {
	t, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}

	if len(t) == 0 {
		return nil, stores.ErrTargetNotFound
	}

	return t[0], nil
}

func (s *InMemoryTargetStore) Save(ctx context.Context, target *models.Target) error {
	tg := *target
	tg.EnvVars = nil
	tg.ApiKey = ""

	s.targets[target.Id] = &tg
	return nil
}

func (s *InMemoryTargetStore) Delete(ctx context.Context, target *models.Target) error {
	delete(s.targets, target.Id)
	return nil
}

func (s *InMemoryTargetStore) processFilters(filter *stores.TargetFilter) ([]*models.Target, error) {
	var result []*models.Target
	filteredTargets := make(map[string]*models.Target)
	for k, v := range s.targets {
		filteredTargets[k] = v
	}

	jobs, err := s.jobMap(context.Background())
	if err != nil {
		return nil, err
	}

	if filter != nil {
		if filter.IdOrName != nil {
			t, ok := s.targets[*filter.IdOrName]
			if ok {
				t.LastJob = jobs[t.Id]
				return []*models.Target{t}, nil
			} else {
				return []*models.Target{}, fmt.Errorf("target with id or name %s not found", *filter.IdOrName)
			}
		}
		if filter.Default != nil {
			for _, t := range filteredTargets {
				if t.IsDefault != *filter.Default {
					delete(filteredTargets, t.Name)
				}
			}
		}
	}

	for _, t := range filteredTargets {
		t.LastJob = jobs[t.Id]
		result = append(result, t)
	}

	return result, nil
}

func (s *InMemoryTargetStore) jobMap(ctx context.Context) (map[string]*models.Job, error) {
	jobs, err := s.jobStore.List(ctx, &stores.JobFilter{
		ResourceType: util.Pointer(models.ResourceTypeTarget),
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
