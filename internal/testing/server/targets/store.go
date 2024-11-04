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

func (s *InMemoryTargetStore) List(filter *target.TargetFilter) ([]*target.Target, error) {
	return s.processFilters(filter)
}

func (s *InMemoryTargetStore) Find(filter *target.TargetFilter) (*target.Target, error) {
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
	s.targets[target.Id] = target
	return nil
}

func (s *InMemoryTargetStore) Delete(target *target.Target) error {
	delete(s.targets, target.Id)
	return nil
}

func (s *InMemoryTargetStore) processFilters(filter *target.TargetFilter) ([]*target.Target, error) {
	var result []*target.Target
	filteredTargets := make(map[string]*target.Target)
	for k, v := range s.targets {
		filteredTargets[k] = v
	}

	if filter != nil {
		if filter.IdOrName != nil {
			t, ok := s.targets[*filter.IdOrName]
			if ok {
				return []*target.Target{t}, nil
			} else {
				return []*target.Target{}, fmt.Errorf("target config with id or name %s not found", *filter.IdOrName)
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
