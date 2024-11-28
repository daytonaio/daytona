//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfigs

import (
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

func (s *InMemoryTargetConfigStore) List(allowDeleted bool) ([]*models.TargetConfig, error) {
	return s.processFilters("", allowDeleted)
}

func (s *InMemoryTargetConfigStore) Find(idOrName string, allowDeleted bool) (*models.TargetConfig, error) {
	targets, err := s.processFilters(idOrName, allowDeleted)
	if err != nil {
		return nil, err
	}

	if len(targets) == 0 {
		return nil, stores.ErrTargetConfigNotFound
	}

	return targets[0], nil
}

func (s *InMemoryTargetConfigStore) Save(targetConfig *models.TargetConfig) error {
	s.targetConfigs[targetConfig.Id] = targetConfig
	return nil
}

func (s *InMemoryTargetConfigStore) processFilters(idOrName string, allowDeleted bool) ([]*models.TargetConfig, error) {
	var result []*models.TargetConfig

	if idOrName != "" {
		t, ok := s.targetConfigs[idOrName]
		if ok {
			result = append(result, t)
		}
	} else {
		for _, targetConfig := range s.targetConfigs {
			result = append(result, targetConfig)
		}
	}

	if !allowDeleted {
		notDeleted := []*models.TargetConfig{}
		for _, targetConfig := range result {
			if !targetConfig.Deleted {
				notDeleted = append(notDeleted, targetConfig)
			}
		}
		result = notDeleted
	}

	return result, nil
}
