// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/daytonaio/daytona/pkg/frpc"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/server/profiledata"
	"github.com/daytonaio/daytona/pkg/server/providertargets"
	"github.com/daytonaio/daytona/pkg/server/registry"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/hashicorp/go-plugin"

	log "github.com/sirupsen/logrus"
)

type ServerInstanceConfig struct {
	Config                   Config
	TailscaleServer          TailscaleServer
	ProviderTargetService    providertargets.IProviderTargetService
	ContainerRegistryService containerregistries.IContainerRegistryService
	LocalContainerRegistry   ILocalContainerRegistry
	WorkspaceService         workspaces.IWorkspaceService
	ApiKeyService            apikeys.IApiKeyService
	GitProviderService       gitproviders.IGitProviderService
	ProviderManager          manager.IProviderManager
	ProfileDataService       profiledata.IProfileDataService
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
			config:                   serverConfig.Config,
			TailscaleServer:          serverConfig.TailscaleServer,
			ProviderTargetService:    serverConfig.ProviderTargetService,
			ContainerRegistryService: serverConfig.ContainerRegistryService,
			LocalContainerRegistry:   serverConfig.LocalContainerRegistry,
			WorkspaceService:         serverConfig.WorkspaceService,
			ApiKeyService:            serverConfig.ApiKeyService,
			GitProviderService:       serverConfig.GitProviderService,
			ProviderManager:          serverConfig.ProviderManager,
			ProfileDataService:       serverConfig.ProfileDataService,
			TelemetryService:         serverConfig.TelemetryService,
		}
	}

	return server
}

type Server struct {
	config                   Config
	TailscaleServer          TailscaleServer
	ProviderTargetService    providertargets.IProviderTargetService
	ContainerRegistryService containerregistries.IContainerRegistryService
	LocalContainerRegistry   ILocalContainerRegistry
	WorkspaceService         workspaces.IWorkspaceService
	ApiKeyService            apikeys.IApiKeyService
	GitProviderService       gitproviders.IGitProviderService
	ProviderManager          manager.IProviderManager
	ProfileDataService       profiledata.IProfileDataService
	TelemetryService         telemetry.TelemetryService
}

func (s *Server) Start(errCh chan error) error {
	err := s.initLogs()
	if err != nil {
		return err
	}

	log.Info("Starting Daytona server")

	headscaleFrpcHealthCheck, headscaleFrpcService, err := frpc.GetService(frpc.FrpcConnectParams{
		ServerDomain: s.config.Frps.Domain,
		ServerPort:   int(s.config.Frps.Port),
		Name:         fmt.Sprintf("daytona-server-%s", s.config.Id),
		Port:         int(s.config.HeadscalePort),
		SubDomain:    s.config.Id,
	})
	if err != nil {
		return err
	}

	if s.LocalContainerRegistry != nil {
		log.Info("Starting local container registry")
		err = s.LocalContainerRegistry.Start()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := registry.RemoveRegistryContainer()
		if err != nil {
			log.Fatalf("Failed to remove local container registry: %s", err.Error())
		}
	}

	go func() {
		err := headscaleFrpcService.Run(context.Background())
		if err != nil {
			errCh <- err
		}
	}()

	apiFrpcHealthCheck, apiFrpcService, err := frpc.GetService(frpc.FrpcConnectParams{
		ServerDomain: s.config.Frps.Domain,
		ServerPort:   int(s.config.Frps.Port),
		Name:         fmt.Sprintf("daytona-server-api-%s", s.config.Id),
		Port:         int(s.config.ApiPort),
		SubDomain:    fmt.Sprintf("api-%s", s.config.Id),
	})
	if err != nil {
		return err
	}

	go func() {
		err := apiFrpcService.Run(context.Background())
		if err != nil {
			errCh <- err
		}
	}()

	if s.LocalContainerRegistry != nil {
		_, registryFrpcService, err := frpc.GetService(frpc.FrpcConnectParams{
			ServerDomain: s.config.Frps.Domain,
			ServerPort:   int(s.config.Frps.Port),
			Name:         fmt.Sprintf("daytona-server-registry-%s", s.config.Id),
			Port:         int(s.config.LocalBuilderRegistryPort),
			SubDomain:    fmt.Sprintf("registry-%s", s.config.Id),
		})
		if err != nil {
			return err
		}
		go func() {
			err := registryFrpcService.Run(context.Background())
			if err != nil {
				errCh <- err
			}
		}()
	}

	go func() {
		interruptChannel := make(chan os.Signal, 1)
		signal.Notify(interruptChannel, os.Interrupt)

		for range interruptChannel {
			log.Info("Shutting down")
			plugin.CleanupClients()
			os.Exit(0)
		}
	}()

	go func() {
		errChan := make(chan error)
		go func() {
			errChan <- s.TailscaleServer.Start()
		}()

		select {
		case err := <-errChan:
			errCh <- err
		case <-time.After(1 * time.Second):
			go func() {
				errChan <- s.TailscaleServer.Connect()
			}()
		}

		if err := <-errChan; err != nil {
			errCh <- err
		}
	}()

	for i := 0; i < 5; i++ {
		if err = headscaleFrpcHealthCheck(); err != nil {
			log.Debugf("Failed to connect to headscale frpc: %s", err)
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}
	if err != nil {
		return err
	}

	for i := 0; i < 5; i++ {
		if err = apiFrpcHealthCheck(); err != nil {
			log.Debugf("Failed to connect to api frpc: %s", err)
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}
	if err != nil {
		return err
	}

	// Terminate orphaned provider processes
	err = s.ProviderManager.TerminateProviderProcesses(s.config.ProvidersDir)
	if err != nil {
		log.Errorf("Failed to terminate orphaned provider processes: %s", err)
	}

	err = s.downloadDefaultProviders()
	if err != nil {
		return err
	}

	err = s.registerProviders()
	if err != nil {
		return err
	}

	return nil
}
