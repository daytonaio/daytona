// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/daytonaio/daytona/common/types"
	"github.com/daytonaio/daytona/server/api"
	"github.com/daytonaio/daytona/server/config"
	"github.com/daytonaio/daytona/server/frpc"
	"github.com/daytonaio/daytona/server/headscale"
	"github.com/daytonaio/daytona/server/logs"
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

	err = downloadDefaultPlugins()
	if err != nil {
		return err
	}

	err = registerProvisioners(c)
	if err != nil {
		return err
	}
	err = registerAgentServices(c)
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

	return api.Start()
}

func StartDaemon() error {

	scriptFile, err := createTemporaryScriptFile()
	if err != nil {
		log.Error(err)
		return nil
	}
	defer func() {
		scriptFile.Close()
		os.Remove(scriptFile.Name())
	}()

	scriptPath := scriptFile.Name()

	// Run the bash script and capture its output
	cmd := exec.Command("bash", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output))
		return err
	}
	log.Info(string(output))

	return nil
}

func getTcpListener(c *types.ServerConfig) (*net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", c.ApiPort))
	if err != nil {
		return nil, err
	}
	return &listener, nil
}
