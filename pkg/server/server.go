// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
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

func Start() error {
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
			log.Fatal(err)
		}
	}()

	go func() {
		if err := frpc.ConnectApi(); err != nil {
			log.Fatal(err)
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
			log.Fatal(err)
		case <-time.After(1 * time.Second):
			go func() {
				errChan <- headscale.Connect()
			}()
		}

		if err := <-errChan; err != nil {
			log.Fatal(err)
		}
	}()

	log.Infof("Starting api server on port %d", c.ApiPort)
	return apiServer.Serve(apiListener)
}
