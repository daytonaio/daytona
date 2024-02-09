// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net"
	"os"
	"os/exec"
	"os/signal"

	proto "github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/daytonaio/daytona/server/config"
	plugin_grpc "github.com/daytonaio/daytona/server/grpc/plugins"
	ports_grpc "github.com/daytonaio/daytona/server/grpc/ports"
	server_grpc "github.com/daytonaio/daytona/server/grpc/server"
	workspace_grpc "github.com/daytonaio/daytona/server/grpc/workspace"
	"github.com/daytonaio/daytona/server/ssh_gateway"
	"github.com/hashicorp/go-plugin"

	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
)

type Self struct {
	HostName string `json:"HostName"`
	DNSName  string `json:"DNSName"`
}

func Start() error {
	log.Info("Starting Daytona server")

	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	_, err = config.GetWorkspaceKey()
	if os.IsNotExist(err) {
		log.Info("Generating workspace key")
		err = config.GenerateWorkspaceKey()
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	lis, err := getUnixListener()
	if err != nil {
		return err
	}
	defer (*lis).Close()

	s := grpc.NewServer()
	workspaceServer := &workspace_grpc.WorkspaceServer{}
	proto.RegisterWorkspaceServiceServer(s, workspaceServer)
	portsServer := &ports_grpc.PortsServer{}
	proto.RegisterPortsServer(s, portsServer)
	serverGrpcServer := &server_grpc.ServerGRPCServer{}
	proto.RegisterServerServer(s, serverGrpcServer)
	pluginsServer := &plugin_grpc.PluginsServer{}
	proto.RegisterPluginsServer(s, pluginsServer)

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

	log.Infof("Daytona server started %v", (*lis).Addr())

	go func() {
		if err := ssh_gateway.Start(); err != nil {
			log.Error(err)
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

	if err := s.Serve(*lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	return nil
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

func getUnixListener() (*net.Listener, error) {
	err := os.RemoveAll("/tmp/daytona/daytona.sock")
	if err != nil {
		return nil, err
	}

	lis, err := net.Listen("unix", "/tmp/daytona/daytona.sock")
	if err != nil {
		return nil, err
	}
	return &lis, nil
}
