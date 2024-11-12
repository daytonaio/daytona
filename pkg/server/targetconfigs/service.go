// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfigs

import (
	"github.com/daytonaio/daytona/pkg/models"
)

type ITargetConfigService interface {
	Delete(targetConfig *models.TargetConfig) error
	Find(filter *TargetConfigFilter) (*models.TargetConfig, error)
	List(filter *TargetConfigFilter) ([]*models.TargetConfig, error)
	Map() (map[string]*models.TargetConfig, error)
	Save(targetConfig *models.TargetConfig) error
}

type TargetConfigServiceConfig struct {
	TargetConfigStore TargetConfigStore
}

type TargetConfigService struct {
	targetConfigStore TargetConfigStore
}

func NewTargetConfigService(config TargetConfigServiceConfig) ITargetConfigService {
	return &TargetConfigService{
		targetConfigStore: config.TargetConfigStore,
	}
}

func (s *TargetConfigService) List(filter *TargetConfigFilter) ([]*models.TargetConfig, error) {
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

func (s *TargetConfigService) Find(filter *TargetConfigFilter) (*models.TargetConfig, error) {
	return s.targetConfigStore.Find(filter)
}

func (s *TargetConfigService) Save(targetConfig *models.TargetConfig) error {
	return s.targetConfigStore.Save(targetConfig)
}

func (s *TargetConfigService) Delete(targetConfig *models.TargetConfig) error {
	return s.targetConfigStore.Delete(targetConfig)
}
