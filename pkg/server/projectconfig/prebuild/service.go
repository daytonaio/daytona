// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/server/projectconfig/prebuild/dto"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
	"github.com/daytonaio/daytona/pkg/workspace/project/config/prebuild"
)

type IPrebuildService interface {
	Find(projectConfigName, id string) (*prebuild.PrebuildConfig, error)
	Set(dto.CreatePrebuildDTO) error
	List(*config.PrebuildFilter) ([]*dto.PrebuildDTO, error)
	Delete(projectConfigName string, prebuild *prebuild.PrebuildConfig) error
}

type PrebuildServiceConfig struct {
	ConfigStore config.Store
}

type PrebuildService struct {
	configStore config.Store
}

func NewPrebuildService(configStore config.Store) IPrebuildService {
	return &PrebuildService{
		configStore: configStore,
	}
}

func (s *PrebuildService) Find(projectConfigName, id string) (*prebuild.PrebuildConfig, error) {
	return nil, nil
}

func (s *PrebuildService) Set(createPrebuildDto dto.CreatePrebuildDTO) error {
	pc, err := s.configStore.Find(&config.Filter{
		Name: &createPrebuildDto.ProjectConfigName,
	})
	if err != nil {
		return config.ErrProjectConfigNotFound
	}

	prebuild := &prebuild.PrebuildConfig{
		Branch:         createPrebuildDto.Branch,
		CommitInterval: createPrebuildDto.CommitInterval,
		TriggerFiles:   createPrebuildDto.TriggerFiles,
	}

	err = prebuild.GenerateId()
	if err != nil {
		return err
	}

	err = pc.SetPrebuild(prebuild)
	if err != nil {
		return err
	}

	err = s.configStore.Save(pc)
	if err != nil {
		return err
	}

	return nil
}

func (s *PrebuildService) List(filter *config.PrebuildFilter) ([]*dto.PrebuildDTO, error) {
	var result []*dto.PrebuildDTO

	if filter == nil {
		pcs, err := s.configStore.List(nil)
		if err != nil {
			return nil, err
		}

		for _, pc := range pcs {
			for _, prebuild := range pc.Prebuilds {
				result = append(result, &dto.PrebuildDTO{
					ProjectConfigName: pc.Name,
					Branch:            prebuild.Branch,
					CommitInterval:    prebuild.CommitInterval,
					TriggerFiles:      prebuild.TriggerFiles,
				})
			}
		}

		fmt.Println(result)

		return result, nil
	}

	pc, err := s.configStore.Find(&config.Filter{
		Name: filter.ProjectConfigName,
	})
	if err != nil {
		return nil, config.ErrProjectConfigNotFound
	}

	for _, prebuild := range pc.Prebuilds {
		result = append(result, &dto.PrebuildDTO{
			ProjectConfigName: pc.Name,
			Branch:            prebuild.Branch,
			CommitInterval:    prebuild.CommitInterval,
			TriggerFiles:      prebuild.TriggerFiles,
		})
	}

	return result, nil
}

func (s *PrebuildService) Delete(projectConfigName string, prebuild *prebuild.PrebuildConfig) error {
	pc, err := s.configStore.Find(&config.Filter{
		Name: &projectConfigName,
	})
	if err != nil {
		return config.ErrProjectConfigNotFound
	}

	err = pc.RemovePrebuild(prebuild)
	if err != nil {
		return err
	}

	return s.configStore.Save(pc)
}
