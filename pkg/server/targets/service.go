// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"io"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type ITargetService interface {
	CreateTarget(ctx context.Context, req dto.CreateTargetDTO) (*target.Target, error)
	GetTarget(ctx context.Context, targetId string, verbose bool) (*dto.TargetDTO, error)
	GetTargetLogReader(targetId string) (io.Reader, error)
	ListTargets(ctx context.Context, verbose bool) ([]dto.TargetDTO, error)
	StartTarget(ctx context.Context, targetId string) error
	StopTarget(ctx context.Context, targetId string) error
	RemoveTarget(ctx context.Context, targetId string) error
	ForceRemoveTarget(ctx context.Context, targetId string) error
}

type targetConfigStore interface {
	Find(filter *provider.TargetConfigFilter) (*provider.TargetConfig, error)
}

type TargetServiceConfig struct {
	TargetStore       target.Store
	TargetConfigStore targetConfigStore
	ServerApiUrl      string
	ServerUrl         string
	ServerVersion     string
	Provisioner       provisioner.IProvisioner
	ApiKeyService     apikeys.IApiKeyService
	LoggerFactory     logs.LoggerFactory
	TelemetryService  telemetry.TelemetryService
}

func NewTargetService(config TargetServiceConfig) ITargetService {
	return &TargetService{
		targetStore:       config.TargetStore,
		targetConfigStore: config.TargetConfigStore,
		serverApiUrl:      config.ServerApiUrl,
		serverUrl:         config.ServerUrl,
		serverVersion:     config.ServerVersion,
		provisioner:       config.Provisioner,
		loggerFactory:     config.LoggerFactory,
		apiKeyService:     config.ApiKeyService,
		telemetryService:  config.TelemetryService,
	}
}

type TargetService struct {
	targetStore       target.Store
	targetConfigStore targetConfigStore
	provisioner       provisioner.IProvisioner
	apiKeyService     apikeys.IApiKeyService
	serverApiUrl      string
	serverUrl         string
	serverVersion     string
	loggerFactory     logs.LoggerFactory
	telemetryService  telemetry.TelemetryService
}

func (s *TargetService) GetTargetLogReader(targetId string) (io.Reader, error) {
	return s.loggerFactory.CreateTargetLogReader(targetId)
}
