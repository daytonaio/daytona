// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfigs

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/docker/docker/pkg/stringid"

	log "github.com/sirupsen/logrus"
)

type TargetConfigServiceConfig struct {
	TargetConfigStore   stores.TargetConfigStore
	TrackTelemetryEvent func(event telemetry.Event, clientId string) error
}

type TargetConfigService struct {
	targetConfigStore   stores.TargetConfigStore
	trackTelemetryEvent func(event telemetry.Event, clientId string) error
}

func NewTargetConfigService(config TargetConfigServiceConfig) services.ITargetConfigService {
	return &TargetConfigService{
		targetConfigStore:   config.TargetConfigStore,
		trackTelemetryEvent: config.TrackTelemetryEvent,
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
		return nil, s.handleCreateError(ctx, nil, err)
	}
	if persistedTargetConfig != nil && !persistedTargetConfig.Deleted {
		return nil, s.handleCreateError(ctx, nil, stores.ErrTargetConfigAlreadyExists)
	}

	targetConfig := &models.TargetConfig{
		Id:           stringid.GenerateRandomID(),
		Name:         addTargetConfig.Name,
		ProviderInfo: addTargetConfig.ProviderInfo,
		Options:      addTargetConfig.Options,
		Deleted:      false,
	}

	err = s.targetConfigStore.Save(ctx, targetConfig)
	return targetConfig, s.handleCreateError(ctx, targetConfig, err)
}

func (s *TargetConfigService) Delete(ctx context.Context, targetConfigId string) error {
	targetConfig, err := s.targetConfigStore.Find(ctx, targetConfigId, false)
	if err != nil {
		return s.handleDeleteError(ctx, nil, err)
	}
	targetConfig.Deleted = true

	err = s.targetConfigStore.Save(ctx, targetConfig)
	return s.handleDeleteError(ctx, targetConfig, err)
}

func (s *TargetConfigService) handleCreateError(ctx context.Context, tc *models.TargetConfig, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	eventName := telemetry.TargetConfigEventLifecycleCreated
	if err != nil {
		eventName = telemetry.TargetConfigEventLifecycleCreationFailed
	}
	event := telemetry.NewTargetConfigEvent(eventName, tc, err, nil)

	telemetryError := s.trackTelemetryEvent(event, clientId)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}

func (s *TargetConfigService) handleDeleteError(ctx context.Context, tc *models.TargetConfig, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	eventName := telemetry.TargetConfigEventLifecycleDeleted
	if err != nil {
		eventName = telemetry.TargetConfigEventLifecycleDeletionFailed
	}
	event := telemetry.NewTargetConfigEvent(eventName, tc, err, nil)

	telemetryError := s.trackTelemetryEvent(event, clientId)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
