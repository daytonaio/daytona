//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/build"
)

type InMemoryBuildStore struct {
	builds map[string]*build.Build
}

func NewInMemoryBuildStore() build.Store {
	return &InMemoryBuildStore{
		builds: make(map[string]*build.Build),
	}
}

func (s *InMemoryBuildStore) Find(filter *build.Filter) (*build.Build, error) {
	builds, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}
	if len(builds) == 0 {
		return nil, build.ErrBuildNotFound
	}

	return builds[0], nil
}

func (s *InMemoryBuildStore) List(filter *build.Filter) ([]*build.Build, error) {
	builds, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}

	return builds, nil
}

func (s *InMemoryBuildStore) Save(result *build.Build) error {
	s.builds[result.Id] = result
	return nil
}

func (s *InMemoryBuildStore) Delete(id string) error {
	delete(s.builds, id)
	return nil
}

func (s *InMemoryBuildStore) processFilters(filter *build.Filter) ([]*build.Build, error) {
	var result []*build.Build
	filteredBuilds := make(map[string]*build.Build)
	for k, v := range s.builds {
		filteredBuilds[k] = v
	}

	if filter != nil {
		if filter.Id != nil {
			b, ok := s.builds[*filter.Id]
			if ok {
				return []*build.Build{b}, nil
			} else {
				return []*build.Build{}, fmt.Errorf("build with id %s not found", *filter.Id)
			}
		}
		if filter.States != nil {
			for _, b := range filteredBuilds {
				check := false
				for _, state := range *filter.States {
					if b.State == state {
						check = true
						break
					}
				}
				if !check {
					delete(filteredBuilds, b.Id)
				}
			}
		}
		if filter.PrebuildIds != nil {
			for _, b := range filteredBuilds {
				check := false
				for _, prebuildId := range *filter.PrebuildIds {
					if b.PrebuildId == prebuildId {
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
			var newestBuild *build.Build
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
				return []*build.Build{newestBuild}, nil
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
				if b.Repository == nil || b.Repository.Branch == nil || *b.Repository.Branch != *filter.Branch {
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
		result = append(result, b)
	}

	return result, nil
}
