//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfigs

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/target/config"
)

type InMemoryTargetConfigStore struct {
	targetConfigs map[string]*config.TargetConfig
}

func NewInMemoryTargetConfigStore() config.TargetConfigStore {
	return &InMemoryTargetConfigStore{
		targetConfigs: make(map[string]*config.TargetConfig),
	}
}

func (s *InMemoryTargetConfigStore) List(filter *config.TargetConfigFilter) ([]*config.TargetConfig, error) {
	return s.processFilters(filter)
}

func (s *InMemoryTargetConfigStore) Find(filter *config.TargetConfigFilter) (*config.TargetConfig, error) {
	targetConfigs, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}
	if len(targetConfigs) == 0 {
		return nil, config.ErrTargetConfigNotFound
	}

	return targetConfigs[0], nil
}

func (s *InMemoryTargetConfigStore) Save(targetConfig *config.TargetConfig) error {
	s.targetConfigs[targetConfig.Name] = targetConfig
	return nil
}

func (s *InMemoryTargetConfigStore) Delete(targetConfig *config.TargetConfig) error {
	delete(s.targetConfigs, targetConfig.Name)
	return nil
}

func (s *InMemoryTargetConfigStore) processFilters(filter *config.TargetConfigFilter) ([]*config.TargetConfig, error) {
	var result []*config.TargetConfig
	targetConfigs := make(map[string]*config.TargetConfig)
	for k, v := range s.targetConfigs {
		targetConfigs[k] = v
	}

	if filter != nil {
		if filter.Name != nil {
			targetConfig, ok := s.targetConfigs[*filter.Name]
			if ok {
				return []*config.TargetConfig{targetConfig}, nil
			} else {
				return []*config.TargetConfig{}, fmt.Errorf("target config with name %s not found", *filter.Name)
			}
		}
	}

	for _, targetConfig := range targetConfigs {
		result = append(result, targetConfig)
	}

	return result, nil
}
