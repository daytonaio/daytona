// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path"

	proto "github.com/daytonaio/daytona/common/grpc/proto"
	agent_service_manager "github.com/daytonaio/daytona/plugins/agent_service/manager"
	provisioner_manager "github.com/daytonaio/daytona/plugins/provisioner/manager"
	"github.com/daytonaio/daytona/server/config"
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
	serverServer := &server_grpc.ServerGRPCServer{}
	proto.RegisterServerServer(s, serverServer)

	err = registerProvisioners(c)
	if err != nil {
		return err
	}
	err = registerServerServices(c)
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

func registerProvisioners(c *config.Config) error {
	provisionerPluginsPath := path.Join(c.PluginsDir, "provisioners")

	files, err := os.ReadDir(provisionerPluginsPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			pluginPath, err := getPluginPath(path.Join(provisionerPluginsPath, file.Name()))
			if err != nil {
				return err
			}

			err = provisioner_manager.RegisterProvisioner(pluginPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func registerServerServices(c *config.Config) error {
	projectServerPluginsPath := path.Join(c.PluginsDir, "server_services")

	files, err := os.ReadDir(projectServerPluginsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	for _, file := range files {
		if file.IsDir() {
			pluginPath, err := getPluginPath(path.Join(projectServerPluginsPath, file.Name()))
			if err != nil {
				return err
			}

			err = agent_service_manager.RegisterAgentService(pluginPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getPluginPath(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if !file.IsDir() {
			return path.Join(dir, file.Name()), nil
		}
	}

	return "", nil
}
