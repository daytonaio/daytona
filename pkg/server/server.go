// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/daytonaio/daytona/pkg/frpc"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/server/api"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/server/providertargets"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
	"github.com/hashicorp/go-plugin"

	log "github.com/sirupsen/logrus"
)

type ServerConfig struct {
	Config                Config
	TailscaleServer       TailscaleServer
	ProviderTargetService providertargets.ProviderTargetService
	WorkspaceService      workspaces.WorkspaceService
	APiKeyService         apikeys.ApiKeyService
	GitProviderService    gitproviders.GitProviderService
}

var server *Server

func GetInstance(serverConfig *ServerConfig) *Server {
	if serverConfig != nil && server != nil {
		log.Fatal("Server already initialized")
	}

	if server == nil {
		server = &Server{
			config:                serverConfig.Config,
			tailscaleServer:       serverConfig.TailscaleServer,
			ProviderTargetService: serverConfig.ProviderTargetService,
			WorkspaceService:      serverConfig.WorkspaceService,
			APiKeyService:         serverConfig.APiKeyService,
			GitProviderService:    serverConfig.GitProviderService,
		}
	}

	return server
}

type Server struct {
	config                Config
	tailscaleServer       TailscaleServer
	ProviderTargetService providertargets.ProviderTargetService
	WorkspaceService      workspaces.WorkspaceService
	APiKeyService         apikeys.ApiKeyService
	GitProviderService    gitproviders.GitProviderService
}

func (s *Server) Start(errCh chan error) error {
	err := s.initLogs()
	if err != nil {
		return err
	}

	log.Info("Starting Daytona server")

	apiServer, err := api.GetServer(int(s.config.ApiPort))
	if err != nil {
		return err
	}

	apiListener, err := net.Listen("tcp", apiServer.Addr)
	if err != nil {
		return err
	}

	// Terminate orphaned provider processes
	err = manager.TerminateProviderProcesses(s.config.ProvidersDir)
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

	go func() {
		err := frpc.Connect(frpc.FrpcConnectParams{
			ServerDomain: s.config.Frps.Domain,
			ServerPort:   int(s.config.Frps.Port),
			Name:         fmt.Sprintf("daytona-server-%s", s.config.Id),
			Port:         int(s.config.HeadscalePort),
			SubDomain:    s.config.Id,
		})
		if err != nil {
			errCh <- err
		}
	}()

	go func() {
		err := frpc.Connect(frpc.FrpcConnectParams{
			ServerDomain: s.config.Frps.Domain,
			ServerPort:   int(s.config.Frps.Port),
			Name:         fmt.Sprintf("daytona-server-api-%s", s.config.Id),
			Port:         int(s.config.ApiPort),
			SubDomain:    fmt.Sprintf("api-%s", s.config.Id),
		})
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
			errChan <- s.tailscaleServer.Start()
		}()

		select {
		case err := <-errChan:
			errCh <- err
		case <-time.After(1 * time.Second):
			go func() {
				errChan <- s.tailscaleServer.Connect()
			}()
		}

		if err := <-errChan; err != nil {
			errCh <- err
		}
	}()

	go func() {
		log.Infof("Starting api server on port %d", s.config.ApiPort)
		err := apiServer.Serve(apiListener)
		if err != nil {
			errCh <- err
		}
	}()

	return nil
}

func (s *Server) HealthCheck() error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf(":%d", s.config.ApiPort), 3*time.Second)
	if err != nil {
		return fmt.Errorf("API health check timed out")
	}
	defer conn.Close()

	return nil
}
