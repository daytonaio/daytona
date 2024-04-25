// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/frpc"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/server/providertargets"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
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
			ServerInstanceConfig: *serverConfig,
		}
	}

	return server
}

type Server struct {
	ServerInstanceConfig
}

func (s *Server) Start(errCh chan error) error {
	err := s.initLogs()
	if err != nil {
		return err
	}

	log.Info("Starting Daytona server")

	// Terminate orphaned provider processes
	err = s.ProviderManager.TerminateProviderProcesses(s.Config.ProvidersDir)
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

	headscaleFrpcHealthCheck, headscaleFrpcService, err := frpc.GetService(frpc.FrpcConnectParams{
		ServerDomain: s.Config.Frps.Domain,
		ServerPort:   int(s.Config.Frps.Port),
		Name:         fmt.Sprintf("daytona-server-%s", s.Config.Id),
		Port:         int(s.Config.HeadscalePort),
		SubDomain:    s.Config.Id,
	})
	if err != nil {
		return err
	}

	//	todo: from config - allow to skip
	log.Info("Starting local container registry")
	err = s.LocalContainerRegistry.Start()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := headscaleFrpcService.Run(context.Background())
		if err != nil {
			errCh <- err
		}
	}()

	apiFrpcHealthCheck, apiFrpcService, err := frpc.GetService(frpc.FrpcConnectParams{
		ServerDomain: s.Config.Frps.Domain,
		ServerPort:   int(s.Config.Frps.Port),
		Name:         fmt.Sprintf("daytona-server-api-%s", s.Config.Id),
		Port:         int(s.Config.ApiPort),
		SubDomain:    fmt.Sprintf("api-%s", s.Config.Id),
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

	_, registryFrpcService, err := frpc.GetService(frpc.FrpcConnectParams{
		ServerDomain: s.Config.Frps.Domain,
		ServerPort:   int(s.Config.Frps.Port),
		Name:         fmt.Sprintf("daytona-server-registry-%s", s.Config.Id),
		Port:         int(s.Config.RegistryPort),
		SubDomain:    fmt.Sprintf("registry-%s", s.Config.Id),
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

	return nil
}

func (s *Server) GetApiUrl() string {
	return util.GetFrpcApiUrl(s.Config.Frps.Protocol, s.Config.Id, s.Config.Frps.Domain)
}
