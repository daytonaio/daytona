//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfigs

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type InMemoryTargetConfigStore struct {
	targetConfigs map[string]*models.TargetConfig
}

func NewInMemoryTargetConfigStore() stores.TargetConfigStore {
	return &InMemoryTargetConfigStore{
		targetConfigs: make(map[string]*models.TargetConfig),
	}
}

func (s *InMemoryTargetConfigStore) List(filter *stores.TargetConfigFilter) ([]*models.TargetConfig, error) {
	return s.processFilters(filter)
}

func (s *InMemoryTargetConfigStore) Find(filter *stores.TargetConfigFilter) (*models.TargetConfig, error) {
	targets, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}

	if len(targets) == 0 {
		return nil, stores.ErrTargetConfigNotFound
	}

	return targets[0], nil
}

func (s *InMemoryTargetConfigStore) Save(targetConfig *models.TargetConfig) error {
	s.targetConfigs[targetConfig.Name] = targetConfig
	return nil
}

func (s *InMemoryTargetConfigStore) Delete(targetConfig *models.TargetConfig) error {
	delete(s.targetConfigs, targetConfig.Name)
	return nil
}

func (s *InMemoryTargetConfigStore) processFilters(filter *stores.TargetConfigFilter) ([]*models.TargetConfig, error) {
	var result []*models.TargetConfig

	if filter != nil {
		if filter.Name != nil {
			t, ok := s.targetConfigs[*filter.Name]
			if ok {
				return []*models.TargetConfig{t}, nil
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
