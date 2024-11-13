// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfigs

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
)

type TargetConfigServiceConfig struct {
	TargetConfigStore stores.TargetConfigStore
}

type TargetConfigService struct {
	targetConfigStore stores.TargetConfigStore
}

func NewTargetConfigService(config TargetConfigServiceConfig) services.ITargetConfigService {
	return &TargetConfigService{
		targetConfigStore: config.TargetConfigStore,
	}
}

func (s *TargetConfigService) List(filter *stores.TargetConfigFilter) ([]*models.TargetConfig, error) {
	return s.targetConfigStore.List(filter)
}

func (s *TargetConfigService) Map() (map[string]*models.TargetConfig, error) {
	list, err := s.targetConfigStore.List(nil)
	if err != nil {
		return nil, err
	}

	targetConfigs := make(map[string]*models.TargetConfig)
	for _, targetConfig := range list {
		targetConfigs[targetConfig.Name] = targetConfig
	}

	return targetConfigs, nil
}

func (s *TargetConfigService) Find(filter *stores.TargetConfigFilter) (*models.TargetConfig, error) {
	return s.targetConfigStore.Find(filter)
}

func (s *TargetConfigService) Save(targetConfig *models.TargetConfig) error {
	return s.targetConfigStore.Save(targetConfig)
}

func (s *TargetConfigService) Delete(targetConfig *models.TargetConfig) error {
	return s.targetConfigStore.Delete(targetConfig)
}
