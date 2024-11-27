// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"io"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provisioner"
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
	CreateJob        func(ctx context.Context, targetId string, action models.JobAction) error

	ServerApiUrl     string
	ServerUrl        string
	ServerVersion    string
	Provisioner      provisioner.IProvisioner
	LoggerFactory    logs.LoggerFactory
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
		provisioner:      config.Provisioner,
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
	createJob        func(ctx context.Context, targetId string, action models.JobAction) error

	provisioner      provisioner.IProvisioner
	serverApiUrl     string
	serverUrl        string
	serverVersion    string
	loggerFactory    logs.LoggerFactory
	telemetryService telemetry.TelemetryService
}

func (s *TargetService) GetTargetLogReader(targetId string) (io.Reader, error) {
	return s.loggerFactory.CreateTargetLogReader(targetId)
}
