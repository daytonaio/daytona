//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/provider"
)

type InMemoryTargetStore struct {
	targets map[string]*provider.ProviderTarget
}

func NewInMemoryTargetStore() provider.TargetStore {
	return &InMemoryTargetStore{
		targets: make(map[string]*provider.ProviderTarget),
	}
}

func (s *InMemoryTargetStore) List(filter *provider.TargetFilter) ([]*provider.ProviderTarget, error) {
	return s.processFilters(filter)
}

func (s *InMemoryTargetStore) Find(filter *provider.TargetFilter) (*provider.ProviderTarget, error) {
	targets, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, provider.ErrTargetNotFound
	}

	return targets[0], nil
}

func (s *InMemoryTargetStore) Save(target *provider.ProviderTarget) error {
	s.targets[target.Name] = target
	return nil
}

func (s *InMemoryTargetStore) Delete(target *provider.ProviderTarget) error {
	delete(s.targets, target.Name)
	return nil
}

func (s *InMemoryTargetStore) processFilters(filter *provider.TargetFilter) ([]*provider.ProviderTarget, error) {
	var result []*provider.ProviderTarget
	filteredTargets := make(map[string]*provider.ProviderTarget)
	for k, v := range s.targets {
		filteredTargets[k] = v
	}

	if filter != nil {
		if filter.Name != nil {
			target, ok := s.targets[*filter.Name]
			if ok {
				return []*provider.ProviderTarget{target}, nil
			} else {
				return []*provider.ProviderTarget{}, fmt.Errorf("target with name %s not found", *filter.Name)
			}
		}
		if filter.Default != nil {
			for _, target := range filteredTargets {
				if target.IsDefault != *filter.Default {
					delete(filteredTargets, target.Name)
				}
			}
		}
	}

	for _, target := range filteredTargets {
		result = append(result, target)
	}

	return result, nil
}
