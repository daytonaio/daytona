// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"reflect"

	. "github.com/daytonaio/daytona/pkg/prebuild"
)

type IPrebuildService interface {
	Find(key string) (*Prebuild, error)
	Set(prebuild *Prebuild) error
	List(filter *PrebuildFilter) ([]*Prebuild, error)
	Delete(prebuild *Prebuild) error
}

type PrebuildServiceConfig struct {
	PrebuildStore ConfigStore
}

func NewPrebuildService(config PrebuildServiceConfig) IPrebuildService {
	return &PrebuildService{
		prebuildStore: config.PrebuildStore,
	}
}

type PrebuildService struct {
	prebuildStore ConfigStore
}

func (s *PrebuildService) Find(id string) (*Prebuild, error) {
	return s.prebuildStore.Find(id)
}

func (s *PrebuildService) Set(prebuild *Prebuild) error {
	persistedPrebuild, err := s.prebuildStore.Find(prebuild.Key)
	if err != nil {
		if IsPrebuildNotFound(err) {
			return s.prebuildStore.Save(prebuild)
		}
		return err
	}

	if prebuild.Branch != "" {
		persistedPrebuild.Branch = prebuild.Branch
	}

	if prebuild.CommitInterval != nil {
		persistedPrebuild.CommitInterval = prebuild.CommitInterval
	}

	if !reflect.DeepEqual(prebuild.TriggerFiles, persistedPrebuild.TriggerFiles) {
		persistedPrebuild.TriggerFiles = prebuild.TriggerFiles
	}

	return s.prebuildStore.Save(persistedPrebuild)
}

func (s *PrebuildService) List(filter *PrebuildFilter) ([]*Prebuild, error) {
	return s.prebuildStore.List(filter)
}

func (s *PrebuildService) Delete(prebuild *Prebuild) error {
	return s.prebuildStore.Delete(prebuild)
}
