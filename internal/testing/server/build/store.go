//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import "github.com/daytonaio/daytona/pkg/build"

type InMemoryBuildStore struct {
	buildResults map[string]*build.BuildResult
}

func NewInMemoryBuildStore() build.Store {
	return &InMemoryBuildStore{
		buildResults: make(map[string]*build.BuildResult),
	}
}

func (s *InMemoryBuildStore) List() ([]*build.BuildResult, error) {
	buildResults := []*build.BuildResult{}
	for _, a := range s.buildResults {
		buildResults = append(buildResults, a)
	}

	return buildResults, nil
}

func (s *InMemoryBuildStore) Find(hash string) (*build.BuildResult, error) {
	buildResult, ok := s.buildResults[hash]
	if !ok {
		return nil, build.ErrBuildNotFound
	}

	return buildResult, nil
}

func (s *InMemoryBuildStore) Save(buildResult *build.BuildResult) error {
	s.buildResults[buildResult.Hash] = buildResult
	return nil
}

func (s *InMemoryBuildStore) Delete(hash string) error {
	delete(s.buildResults, hash)
	return nil
}
