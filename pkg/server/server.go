// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/server/api"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/frpc"
	"github.com/daytonaio/daytona/pkg/server/headscale"
	"github.com/daytonaio/daytona/pkg/server/logs"
	"github.com/hashicorp/go-plugin"

	log "github.com/sirupsen/logrus"
)

type Self struct {
	HostName string `json:"HostName"`
	DNSName  string `json:"DNSName"`
}

func Start(errCh chan error) error {
	err := logs.Init()
	if err != nil {
		return err
	}

	log.Info("Starting Daytona server")

	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	apiServer, err := api.GetServer()
	if err != nil {
		return err
	}

	apiListener, err := net.Listen("tcp", apiServer.Addr)
	if err != nil {
		return err
	}

	// Terminate orphaned provider processes
	err = manager.TerminateProviderProcesses(c.ProvidersDir)
	if err != nil {
		log.Errorf("Failed to terminate orphaned provider processes: %s", err)
	}

	err = downloadDefaultProviders()
	if err != nil {
		return err
	}

	err = registerProviders(c)
	if err != nil {
		return err
	}

	go func() {
		if err := frpc.ConnectServer(); err != nil {
			errCh <- err
		}
	}()

	go func() {
		if err := frpc.ConnectApi(); err != nil {
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
			errChan <- headscale.Start(c)
		}()

		select {
		case err := <-errChan:
			errCh <- err
		case <-time.After(1 * time.Second):
			go func() {
				errChan <- headscale.Connect()
			}()
		}

		if err := <-errChan; err != nil {
			errCh <- err
		}
	}()

	go func() {
		log.Infof("Starting api server on port %d", c.ApiPort)
		err := apiServer.Serve(apiListener)
		if err != nil {
			errCh <- err
		}
	}()

	return nil
}

func HealthCheck() error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf(":%d", c.ApiPort), 3*time.Second)
	if err != nil {
		return fmt.Errorf("API health check timed out")
	}
	defer conn.Close()

	return nil
}
