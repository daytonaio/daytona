// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfigs

import (
	"github.com/daytonaio/daytona/pkg/target/config"
)

type ITargetConfigService interface {
	Delete(targetConfig *config.TargetConfig) error
	Find(filter *config.TargetConfigFilter) (*config.TargetConfig, error)
	List(filter *config.TargetConfigFilter) ([]*config.TargetConfig, error)
	Map() (map[string]*config.TargetConfig, error)
	Save(targetConfig *config.TargetConfig) error
}

type TargetConfigServiceConfig struct {
	TargetConfigStore config.TargetConfigStore
}

type TargetConfigService struct {
	targetConfigStore config.TargetConfigStore
}

func NewTargetConfigService(config TargetConfigServiceConfig) ITargetConfigService {
	return &TargetConfigService{
		targetConfigStore: config.TargetConfigStore,
	}
}

func (s *TargetConfigService) List(filter *config.TargetConfigFilter) ([]*config.TargetConfig, error) {
	return s.targetConfigStore.List(filter)
}

func (s *TargetConfigService) Map() (map[string]*config.TargetConfig, error) {
	list, err := s.targetConfigStore.List(nil)
	if err != nil {
		return nil, err
	}

	targetConfigs := make(map[string]*config.TargetConfig)
	for _, targetConfig := range list {
		targetConfigs[targetConfig.Name] = targetConfig
	}

	return targetConfigs, nil
}

func (s *TargetConfigService) Find(filter *config.TargetConfigFilter) (*config.TargetConfig, error) {
	return s.targetConfigStore.Find(filter)
}

func (s *TargetConfigService) Save(targetConfig *config.TargetConfig) error {
	return s.targetConfigStore.Save(targetConfig)
}

func (s *TargetConfigService) Delete(targetConfig *config.TargetConfig) error {
	return s.targetConfigStore.Delete(targetConfig)
}
