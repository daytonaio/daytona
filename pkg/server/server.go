// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/telemetry"

	log "github.com/sirupsen/logrus"
)

type ServerInstanceConfig struct {
	Config                     Config
	Version                    string
	TailscaleServer            TailscaleServer
	TargetConfigService        services.ITargetConfigService
	BuildService               services.IBuildService
	WorkspaceTemplateService   services.IWorkspaceTemplateService
	WorkspaceService           services.IWorkspaceService
	LocalContainerRegistry     ILocalContainerRegistry
	TargetService              services.ITargetService
	ApiKeyService              services.IApiKeyService
	GitProviderService         services.IGitProviderService
	EnvironmentVariableService services.IEnvironmentVariableService
	JobService                 services.IJobService
	RunnerService              services.IRunnerService
	TelemetryService           telemetry.TelemetryService
}

var server *Server

func GetInstance(serverConfig *ServerInstanceConfig) *Server {
	if serverConfig != nil && server != nil {
		log.Fatal("Server already initialized")
	}

	if server == nil {
		if serverConfig == nil {
			log.Fatal("Server not initialized")
		}
		server = &Server{
			Id:                         serverConfig.Config.Id,
			config:                     serverConfig.Config,
			Version:                    serverConfig.Version,
			TailscaleServer:            serverConfig.TailscaleServer,
			TargetConfigService:        serverConfig.TargetConfigService,
			BuildService:               serverConfig.BuildService,
			WorkspaceTemplateService:   serverConfig.WorkspaceTemplateService,
			WorkspaceService:           serverConfig.WorkspaceService,
			LocalContainerRegistry:     serverConfig.LocalContainerRegistry,
			TargetService:              serverConfig.TargetService,
			ApiKeyService:              serverConfig.ApiKeyService,
			GitProviderService:         serverConfig.GitProviderService,
			EnvironmentVariableService: serverConfig.EnvironmentVariableService,
			JobService:                 serverConfig.JobService,
			RunnerService:              serverConfig.RunnerService,
			TelemetryService:           serverConfig.TelemetryService,
		}
	}

	return server
}

type Server struct {
	Id                         string
	config                     Config
	Version                    string
	TailscaleServer            TailscaleServer
	TargetConfigService        services.ITargetConfigService
	BuildService               services.IBuildService
	WorkspaceTemplateService   services.IWorkspaceTemplateService
	WorkspaceService           services.IWorkspaceService
	LocalContainerRegistry     ILocalContainerRegistry
	TargetService              services.ITargetService
	ApiKeyService              services.IApiKeyService
	GitProviderService         services.IGitProviderService
	EnvironmentVariableService services.IEnvironmentVariableService
	JobService                 services.IJobService
	RunnerService              services.IRunnerService
	TelemetryService           telemetry.TelemetryService
}

func (s *Server) Initialize() error {
	return s.initLogs()
}
