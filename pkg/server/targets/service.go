// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type TargetServiceConfig struct {
	TargetStore         stores.TargetStore
	TargetMetadataStore stores.TargetMetadataStore

	FindTargetConfig func(ctx context.Context, name string) (*models.TargetConfig, error)
	GenerateApiKey   func(ctx context.Context, name string) (string, error)
	RevokeApiKey     func(ctx context.Context, name string) error
	CreateJob        func(ctx context.Context, targetId string, runnerId string, action models.JobAction) error

	ServerApiUrl     string
	ServerUrl        string
	ServerVersion    string
	LoggerFactory    logs.ILoggerFactory
	TelemetryService telemetry.TelemetryService
}

func NewTargetService(config TargetServiceConfig) services.ITargetService {
	return &TargetService{
		targetStore:         config.TargetStore,
		targetMetadataStore: config.TargetMetadataStore,

		findTargetConfig: config.FindTargetConfig,
		generateApiKey:   config.GenerateApiKey,
		revokeApiKey:     config.RevokeApiKey,
		createJob:        config.CreateJob,

		serverApiUrl:     config.ServerApiUrl,
		serverUrl:        config.ServerUrl,
		serverVersion:    config.ServerVersion,
		loggerFactory:    config.LoggerFactory,
		telemetryService: config.TelemetryService,
	}
}

type TargetService struct {
	targetStore         stores.TargetStore
	targetMetadataStore stores.TargetMetadataStore

	findTargetConfig func(ctx context.Context, name string) (*models.TargetConfig, error)
	generateApiKey   func(ctx context.Context, name string) (string, error)
	revokeApiKey     func(ctx context.Context, name string) error
	createJob        func(ctx context.Context, targetId string, runnerId string, action models.JobAction) error

	serverApiUrl     string
	serverUrl        string
	serverVersion    string
	loggerFactory    logs.ILoggerFactory
	telemetryService telemetry.TelemetryService
}

func (s *TargetService) GetTargetLogReader(targetId string) (io.Reader, error) {
	return s.loggerFactory.CreateTargetLogReader(targetId)
}

func (s *TargetService) GetTargetLogWriter(ctx context.Context, targetId string) (io.WriteCloser, error) {
	configDir, err := server.GetConfigDir()
	if err != nil {
		return nil, err
	}

	targetLogsDir, err := server.GetTargetLogsDir(configDir)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(targetLogsDir, 0755)
	if err != nil {
		return nil, err
	}

	return os.OpenFile(filepath.Join(targetLogsDir, targetId, "log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

// TODO: revise - "remove default" is enough for now
func (s *TargetService) SaveTarget(ctx context.Context, target *models.Target) error {
	return s.targetStore.Save(ctx, target)
}

func (s *TargetService) UpdateTargetProviderMetadata(ctx context.Context, targetId, metadata string) error {
	tg, err := s.targetStore.Find(ctx, &stores.TargetFilter{
		IdOrName: &targetId,
	})
	if err != nil {
		return err
	}

	tg.ProviderMetadata = &metadata
	return s.targetStore.Save(ctx, tg)
}
