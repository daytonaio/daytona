// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfigs

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/provider"
)

type ITargetConfigService interface {
	Delete(targetConfig *provider.TargetConfig) error
	Find(filter *provider.TargetConfigFilter) (*provider.TargetConfig, error)
	List(filter *provider.TargetConfigFilter) ([]*provider.TargetConfig, error)
	Map() (map[string]*provider.TargetConfig, error)
	Save(targetConfig *provider.TargetConfig) error
	SetDefault(targetConfig *provider.TargetConfig) error
}

type TargetConfigServiceConfig struct {
	TargetConfigStore provider.TargetConfigStore
}

type TargetConfigService struct {
	targetConfigStore provider.TargetConfigStore
}

func NewTargetConfigService(config TargetConfigServiceConfig) ITargetConfigService {
	return &TargetConfigService{
		targetConfigStore: config.TargetConfigStore,
	}
}

func (s *TargetConfigService) List(filter *provider.TargetConfigFilter) ([]*provider.TargetConfig, error) {
	return s.targetConfigStore.List(filter)
}

func (s *TargetConfigService) Map() (map[string]*provider.TargetConfig, error) {
	list, err := s.targetConfigStore.List(nil)
	if err != nil {
		return nil, err
	}

	targetConfigs := make(map[string]*provider.TargetConfig)
	for _, targetConfig := range list {
		targetConfigs[targetConfig.Name] = targetConfig
	}

	return targetConfigs, nil
}

func (s *TargetConfigService) Find(filter *provider.TargetConfigFilter) (*provider.TargetConfig, error) {
	return s.targetConfigStore.Find(filter)
}

func (s *TargetConfigService) Save(targetConfig *provider.TargetConfig) error {
	err := s.targetConfigStore.Save(targetConfig)
	if err != nil {
		return err
	}

	return s.SetDefault(targetConfig)
}

func (s *TargetConfigService) Delete(targetConfig *provider.TargetConfig) error {
	return s.targetConfigStore.Delete(targetConfig)
}

func (s *TargetConfigService) SetDefault(targetConfig *provider.TargetConfig) error {
	currentConfig, err := s.Find(&provider.TargetConfigFilter{
		Name: &targetConfig.Name,
	})
	if err != nil {
		return err
	}

	defaultConfig, err := s.Find(&provider.TargetConfigFilter{
		Default: util.Pointer(true),
	})
	if err != nil && err != provider.ErrTargetConfigNotFound {
		return err
	}

	if defaultConfig != nil {
		defaultConfig.IsDefault = false
		err := s.targetConfigStore.Save(defaultConfig)
		if err != nil {
			return err
		}
	}

	currentConfig.IsDefault = true
	return s.targetConfigStore.Save(currentConfig)
}
