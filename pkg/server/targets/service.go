// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"io"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type TargetServiceConfig struct {
	TargetStore         stores.TargetStore
	TargetMetadataStore stores.TargetMetadataStore

	FindTargetConfig    func(ctx context.Context, name string) (*models.TargetConfig, error)
	CreateApiKey        func(ctx context.Context, name string) (string, error)
	DeleteApiKey        func(ctx context.Context, name string) error
	CreateJob           func(ctx context.Context, targetId string, runnerId string, action models.JobAction) error
	TrackTelemetryEvent func(event telemetry.Event, clientId string) error

	ServerApiUrl  string
	ServerUrl     string
	ServerVersion string
	LoggerFactory logs.ILoggerFactory
}

func NewTargetService(config TargetServiceConfig) services.ITargetService {
	return &TargetService{
		targetStore:         config.TargetStore,
		targetMetadataStore: config.TargetMetadataStore,

		findTargetConfig: config.FindTargetConfig,
		createApiKey:     config.CreateApiKey,
		deleteApiKey:     config.DeleteApiKey,
		createJob:        config.CreateJob,

		serverApiUrl:        config.ServerApiUrl,
		serverUrl:           config.ServerUrl,
		serverVersion:       config.ServerVersion,
		loggerFactory:       config.LoggerFactory,
		trackTelemetryEvent: config.TrackTelemetryEvent,
	}
}

type TargetService struct {
	targetStore         stores.TargetStore
	targetMetadataStore stores.TargetMetadataStore

	findTargetConfig    func(ctx context.Context, name string) (*models.TargetConfig, error)
	createApiKey        func(ctx context.Context, name string) (string, error)
	deleteApiKey        func(ctx context.Context, name string) error
	createJob           func(ctx context.Context, targetId string, runnerId string, action models.JobAction) error
	trackTelemetryEvent func(event telemetry.Event, clientId string) error

	serverApiUrl  string
	serverUrl     string
	serverVersion string
	loggerFactory logs.ILoggerFactory
}

func (s *TargetService) GetTargetLogReader(ctx context.Context, targetId string) (io.Reader, error) {
	return s.loggerFactory.CreateLogReader(targetId)
}

func (s *TargetService) GetTargetLogWriter(ctx context.Context, targetId string) (io.WriteCloser, error) {
	return s.loggerFactory.CreateLogWriter(targetId)
}

// TODO: revise - "remove default" is enough for now
func (s *TargetService) Save(ctx context.Context, target *models.Target) error {
	return s.targetStore.Save(ctx, target)
}

func (s *TargetService) UpdateProviderMetadata(ctx context.Context, targetId, metadata string) error {
	tg, err := s.targetStore.Find(ctx, &stores.TargetFilter{
		IdOrName: &targetId,
	})
	if err != nil {
		return err
	}

	tg.ProviderMetadata = &metadata
	return s.targetStore.Save(ctx, tg)
}

func (s *TargetService) UpdateLastJob(ctx context.Context, targetId, jobId string) error {
	t, err := s.targetStore.Find(ctx, &stores.TargetFilter{
		IdOrName: &targetId,
	})
	if err != nil {
		return err
	}

	t.LastJobId = &jobId
	// Make sure the old relation doesn't get saved to the store
	t.LastJob = nil

	return s.targetStore.Save(ctx, t)
}
