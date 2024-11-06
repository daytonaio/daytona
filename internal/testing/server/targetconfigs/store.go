//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfigs

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/target"
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
	targets, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}

	if len(targets) == 0 {
		return nil, target.ErrTargetNotFound
	}

	return targets[0], nil
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

	if filter != nil {
		if filter.Name != nil {
			t, ok := s.targetConfigs[*filter.Name]
			if ok {
				return []*config.TargetConfig{t}, nil
			} else {
				return nil, fmt.Errorf("target config with id or name %s not found", *filter.Name)
			}
		}
	}

	for _, targetConfig := range s.targetConfigs {
		result = append(result, targetConfig)
	}

	return result, nil
}
