// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"os"
	"os/signal"

	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/hashicorp/go-plugin"

	log "github.com/sirupsen/logrus"
)

type ServerInstanceConfig struct {
	Config                   Config
	Version                  string
	TailscaleServer          TailscaleServer
	TargetConfigService      services.ITargetConfigService
	ContainerRegistryService services.IContainerRegistryService
	BuildService             services.IBuildService
	WorkspaceConfigService   services.IWorkspaceConfigService
	WorkspaceService         services.IWorkspaceService
	LocalContainerRegistry   ILocalContainerRegistry
	TargetService            services.ITargetService
	ApiKeyService            services.IApiKeyService
	GitProviderService       services.IGitProviderService
	ProviderManager          manager.IProviderManager
	ProfileDataService       services.IProfileDataService
	JobService               services.IJobService
	TelemetryService         telemetry.TelemetryService
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
			Id:                       serverConfig.Config.Id,
			config:                   serverConfig.Config,
			Version:                  serverConfig.Version,
			TailscaleServer:          serverConfig.TailscaleServer,
			TargetConfigService:      serverConfig.TargetConfigService,
			ContainerRegistryService: serverConfig.ContainerRegistryService,
			BuildService:             serverConfig.BuildService,
			WorkspaceConfigService:   serverConfig.WorkspaceConfigService,
			WorkspaceService:         serverConfig.WorkspaceService,
			LocalContainerRegistry:   serverConfig.LocalContainerRegistry,
			TargetService:            serverConfig.TargetService,
			ApiKeyService:            serverConfig.ApiKeyService,
			GitProviderService:       serverConfig.GitProviderService,
			ProviderManager:          serverConfig.ProviderManager,
			ProfileDataService:       serverConfig.ProfileDataService,
			JobService:               serverConfig.JobService,
			TelemetryService:         serverConfig.TelemetryService,
		}
	}

	return server
}

type Server struct {
	Id                       string
	config                   Config
	Version                  string
	TailscaleServer          TailscaleServer
	TargetConfigService      services.ITargetConfigService
	ContainerRegistryService services.IContainerRegistryService
	BuildService             services.IBuildService
	WorkspaceConfigService   services.IWorkspaceConfigService
	WorkspaceService         services.IWorkspaceService
	LocalContainerRegistry   ILocalContainerRegistry
	TargetService            services.ITargetService
	ApiKeyService            services.IApiKeyService
	GitProviderService       services.IGitProviderService
	ProviderManager          manager.IProviderManager
	ProfileDataService       services.IProfileDataService
	JobService               services.IJobService
	TelemetryService         telemetry.TelemetryService
}

func (s *Server) Initialize() error {
	return s.initLogs()
}

func (s *Server) Start() error {
	log.Info("Starting Daytona server")

	go func() {
		interruptChannel := make(chan os.Signal, 1)
		signal.Notify(interruptChannel, os.Interrupt)

		for range interruptChannel {
			plugin.CleanupClients()
		}
	}()

	// Terminate orphaned provider processes
	err := s.ProviderManager.TerminateProviderProcesses(s.config.ProvidersDir)
	if err != nil {
		log.Errorf("Failed to terminate orphaned provider processes: %s", err)
	}

	err = s.downloadDefaultProviders()
	if err != nil {
		return err
	}

	return s.registerProviders()
}
