// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfigs

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/docker/docker/pkg/stringid"
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

func (s *TargetConfigService) List(ctx context.Context) ([]*models.TargetConfig, error) {
	return s.targetConfigStore.List(ctx, false)
}

func (s *TargetConfigService) Map(ctx context.Context) (map[string]*models.TargetConfig, error) {
	list, err := s.targetConfigStore.List(ctx, false)
	if err != nil {
		return nil, err
	}

	targetConfigs := make(map[string]*models.TargetConfig)
	for _, targetConfig := range list {
		targetConfigs[targetConfig.Name] = targetConfig
	}

	return targetConfigs, nil
}

func (s *TargetConfigService) Find(ctx context.Context, idOrName string) (*models.TargetConfig, error) {
	return s.targetConfigStore.Find(ctx, idOrName, false)
}

func (s *TargetConfigService) Add(ctx context.Context, addTargetConfig services.AddTargetConfigDTO) (*models.TargetConfig, error) {
	persistedTargetConfig, err := s.targetConfigStore.Find(ctx, addTargetConfig.Name, false)
	if err != nil && !stores.IsTargetConfigNotFound(err) {
		return nil, err
	}
	if persistedTargetConfig != nil && !persistedTargetConfig.Deleted {
		return nil, stores.ErrTargetConfigAlreadyExists
	}

	targetConfig := &models.TargetConfig{
		Id:           stringid.GenerateRandomID(),
		Name:         addTargetConfig.Name,
		ProviderInfo: addTargetConfig.ProviderInfo,
		Options:      addTargetConfig.Options,
		Deleted:      false,
	}

	return targetConfig, s.targetConfigStore.Save(ctx, targetConfig)
}

func (s *TargetConfigService) Delete(ctx context.Context, targetConfigId string) error {
	targetConfig, err := s.targetConfigStore.Find(ctx, targetConfigId, false)
	if err != nil {
		return err
	}
	targetConfig.Deleted = true

	return s.targetConfigStore.Save(ctx, targetConfig)
}
