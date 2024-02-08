// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"net"
	"os"
	"os/exec"
	"path"

	"github.com/daytonaio/daytona/agent/config"
	agent_grpc "github.com/daytonaio/daytona/agent/grpc/agent"
	ports_grpc "github.com/daytonaio/daytona/agent/grpc/ports"
	workspace_grpc "github.com/daytonaio/daytona/agent/grpc/workspace"
	"github.com/daytonaio/daytona/agent/ssh_gateway"
	proto "github.com/daytonaio/daytona/grpc/proto"
	project_agent_manager "github.com/daytonaio/daytona/plugin/project_agent/manager"
	provisioner_manager "github.com/daytonaio/daytona/plugin/provisioner/manager"

	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
)

type Self struct {
	HostName string `json:"HostName"`
	DNSName  string `json:"DNSName"`
}

func Start() error {
	log.Info("Starting Daytona agent")

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
	agentServer := &agent_grpc.AgentServer{}
	proto.RegisterAgentServer(s, agentServer)

	err = registerProvisioners(c)
	if err != nil {
		return err
	}
	err = registerProjectAgents(c)
	if err != nil {
		return err
	}

	log.Infof("Daytona agent started %v", (*lis).Addr())

	go func() {
		if err := ssh_gateway.Start(); err != nil {
			log.Error(err)
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
		if !file.IsDir() {
			err := provisioner_manager.RegisterProvisioner(path.Join(provisionerPluginsPath, file.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func registerProjectAgents(c *config.Config) error {
	projectAgentPluginsPath := path.Join(c.PluginsDir, "project_agents")

	files, err := os.ReadDir(projectAgentPluginsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			err := project_agent_manager.RegisterProjectAgent(path.Join(projectAgentPluginsPath, file.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
