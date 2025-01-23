//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/testing/common"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type InMemoryBuildStore struct {
	common.InMemoryStore
	builds   map[string]*models.Build
	jobStore stores.JobStore
}

func NewInMemoryBuildStore(jobStore stores.JobStore) stores.BuildStore {
	return &InMemoryBuildStore{
		builds:   make(map[string]*models.Build),
		jobStore: jobStore,
	}
}

func (s *InMemoryBuildStore) Find(ctx context.Context, filter *stores.BuildFilter) (*models.Build, error) {
	b, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, stores.ErrBuildNotFound
	}

	return b[0], nil
}

func (s *InMemoryBuildStore) List(ctx context.Context, filter *stores.BuildFilter) ([]*models.Build, error) {
	builds, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}

	return builds, nil
}

func (s *InMemoryBuildStore) Save(ctx context.Context, result *models.Build) error {
	s.builds[result.Id] = result
	return nil
}

func (s *InMemoryBuildStore) Delete(ctx context.Context, id string) error {
	delete(s.builds, id)
	return nil
}

func (s *InMemoryBuildStore) processFilters(filter *stores.BuildFilter) ([]*models.Build, error) {
	var result []*models.Build
	filteredBuilds := make(map[string]*models.Build)
	for k, v := range s.builds {
		filteredBuilds[k] = v
	}

	jobs, err := s.jobMap(context.Background())
	if err != nil {
		return nil, err
	}

	if filter != nil {
		if filter.Id != nil {
			b, ok := s.builds[*filter.Id]
			if ok {
				b.LastJob = jobs[b.Id]
				return []*models.Build{b}, nil
			} else {
				return []*models.Build{}, fmt.Errorf("build with id %s not found", *filter.Id)
			}
		}
		if filter.PrebuildIds != nil {
			for _, b := range filteredBuilds {
				check := false
				for _, prebuildId := range *filter.PrebuildIds {
					if b.PrebuildId != nil && *b.PrebuildId == prebuildId {
						check = true
						break
					}
				}
				if !check {
					delete(filteredBuilds, b.Id)
				}
			}
		}
		if filter.GetNewest != nil && *filter.GetNewest {
			var newestBuild *models.Build
			for _, b := range filteredBuilds {
				if newestBuild == nil {
					newestBuild = b
					continue
				}
				if b.CreatedAt.After(newestBuild.CreatedAt) {
					newestBuild = b
				}
			}
			if newestBuild != nil {
				newestBuild.LastJob = jobs[newestBuild.Id]
				return []*models.Build{newestBuild}, nil
			}
		}
		if filter.BuildConfig != nil {
			for _, b := range filteredBuilds {
				if b.BuildConfig == nil || b.BuildConfig != filter.BuildConfig {
					delete(filteredBuilds, b.Id)
				}
			}
		}
		if filter.RepositoryUrl != nil {
			for _, b := range filteredBuilds {
				if b.Repository == nil || b.Repository.Url != *filter.RepositoryUrl {
					delete(filteredBuilds, b.Id)
				}
			}
		}
		if filter.Branch != nil {
			for _, b := range filteredBuilds {
				if b.Repository == nil || b.Repository.Branch != *filter.Branch {
					delete(filteredBuilds, b.Id)
				}
			}
		}
		if filter.EnvVars != nil {
			for _, b := range filteredBuilds {
				if b.EnvVars == nil {
					delete(filteredBuilds, b.Id)
					continue
				}
				for key, value := range *filter.EnvVars {
					if b.EnvVars[key] != value {
						delete(filteredBuilds, b.Id)
						break
					}
				}
			}
		}
	}

	for _, b := range filteredBuilds {
		b.LastJob = jobs[b.Id]
		result = append(result, b)
	}

	return result, nil
}

func (s *InMemoryBuildStore) jobMap(ctx context.Context) (map[string]*models.Job, error) {
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
