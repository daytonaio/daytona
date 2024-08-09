// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package builds

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/build"
)

type IBuildService interface {
	Create(*build.Build) error
	Find(id string) (*build.Build, error)
	List(filter *build.BuildFilter) ([]*build.Build, error)
	Delete(id string) error
}

type BuildServiceConfig struct {
	BuildStore build.Store
}

type BuildService struct {
	buildStore build.Store
}

func NewBuildService(config BuildServiceConfig) IBuildService {
	return &BuildService{
		buildStore: config.BuildStore,
	}
}

func (s *BuildService) Create(b *build.Build) error {
	b.Hash = "test"
	b.Id = "asdf"
	b.State = build.BuildStatePending
	return s.buildStore.Save(b)
}

func (s *BuildService) Find(id string) (*build.Build, error) {
	return s.buildStore.Find(id)
}

func (s *BuildService) List(filter *build.BuildFilter) ([]*build.Build, error) {
	result, err := s.buildStore.List(filter)
	if err != nil {
		return nil, err
	}

	fmt.Println(result[0])

	return result, nil
}

func (s *BuildService) Delete(id string) error {
	return s.buildStore.Delete(id)
}
