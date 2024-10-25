//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfigs

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/provider"
)

type InMemoryTargetConfigStore struct {
	targetConfigs map[string]*provider.TargetConfig
}

func NewInMemoryTargetConfigStore() provider.TargetConfigStore {
	return &InMemoryTargetConfigStore{
		targetConfigs: make(map[string]*provider.TargetConfig),
	}
}

func (s *InMemoryTargetConfigStore) List(filter *provider.TargetConfigFilter) ([]*provider.TargetConfig, error) {
	return s.processFilters(filter)
}

func (s *InMemoryTargetConfigStore) Find(filter *provider.TargetConfigFilter) (*provider.TargetConfig, error) {
	targetConfigs, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}
	if len(targetConfigs) == 0 {
		return nil, provider.ErrTargetConfigNotFound
	}

	return targetConfigs[0], nil
}

func (s *InMemoryTargetConfigStore) Save(targetConfig *provider.TargetConfig) error {
	s.targetConfigs[targetConfig.Name] = targetConfig
	return nil
}

func (s *InMemoryTargetConfigStore) Delete(targetConfig *provider.TargetConfig) error {
	delete(s.targetConfigs, targetConfig.Name)
	return nil
}

func (s *InMemoryTargetConfigStore) processFilters(filter *provider.TargetConfigFilter) ([]*provider.TargetConfig, error) {
	var result []*provider.TargetConfig
	targetConfigs := make(map[string]*provider.TargetConfig)
	for k, v := range s.targetConfigs {
		targetConfigs[k] = v
	}

	if filter != nil {
		if filter.Name != nil {
			targetConfig, ok := s.targetConfigs[*filter.Name]
			if ok {
				return []*provider.TargetConfig{targetConfig}, nil
			} else {
				return []*provider.TargetConfig{}, fmt.Errorf("target config with name %s not found", *filter.Name)
			}
		}
		if filter.Default != nil {
			for _, targetConfig := range targetConfigs {
				if targetConfig.IsDefault != *filter.Default {
					delete(targetConfigs, targetConfig.Name)
				}
			}
		}
	}

	for _, targetConfig := range targetConfigs {
		result = append(result, targetConfig)
	}

	return result, nil
}
