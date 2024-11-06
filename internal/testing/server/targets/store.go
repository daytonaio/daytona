//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/target"
)

type InMemoryTargetStore struct {
	targets map[string]*target.Target
}

func NewInMemoryTargetStore() target.Store {
	return &InMemoryTargetStore{
		targets: make(map[string]*target.Target),
	}
}

func (s *InMemoryTargetStore) List(filter *target.TargetFilter) ([]*target.TargetViewDTO, error) {
	return s.processFilters(filter)
}

func (s *InMemoryTargetStore) Find(filter *target.TargetFilter) (*target.TargetViewDTO, error) {
	targets, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}

	if len(targets) == 0 {
		return nil, target.ErrTargetNotFound
	}

	return targets[0], nil
}

func (s *InMemoryTargetStore) Save(target *target.Target) error {
	tg := *target
	tg.EnvVars = nil
	tg.ApiKey = ""

	s.targets[target.Id] = &tg
	return nil
}

func (s *InMemoryTargetStore) Delete(target *target.Target) error {
	delete(s.targets, target.Id)
	return nil
}

func (s *InMemoryTargetStore) processFilters(filter *target.TargetFilter) ([]*target.TargetViewDTO, error) {
	var result []*target.TargetViewDTO
	filteredTargets := make(map[string]*target.TargetViewDTO)
	for k, v := range s.targets {
		filteredTargets[k] = &target.TargetViewDTO{
			Target: *v,
		}
	}

	if filter != nil {
		if filter.IdOrName != nil {
			t, ok := s.targets[*filter.IdOrName]
			if ok {
				return []*target.TargetViewDTO{{Target: *t}}, nil
			} else {
				return []*target.TargetViewDTO{}, fmt.Errorf("target with id or name %s not found", *filter.IdOrName)
			}
		}
		if filter.Default != nil {
			for _, targetConfig := range filteredTargets {
				if targetConfig.IsDefault != *filter.Default {
					delete(filteredTargets, targetConfig.Name)
				}
			}
		}
	}

	for _, targetConfig := range filteredTargets {
		result = append(result, targetConfig)
	}

	return result, nil
}
