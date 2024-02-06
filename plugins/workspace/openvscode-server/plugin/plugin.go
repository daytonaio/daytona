// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/daytonaio/daytona/agent/workspace"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/plugin"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type OpenVSCodeServerPlugin struct {
	BasePath string
}

var Config *plugin.WorkspacePluginConfig

func (e OpenVSCodeServerPlugin) SetConfig(config plugin.WorkspacePluginConfig) error {
	Config = &config
	return nil
}

func (e OpenVSCodeServerPlugin) GetName() string {
	return "openvscode-server"
}

func (e OpenVSCodeServerPlugin) GetVersion() string {
	return "0.0.1"
}

func (e OpenVSCodeServerPlugin) ProjectPreInit(project workspace.Project) error {
	setupDir := Config.SetupPath

	err := os.MkdirAll(path.Join(setupDir, "server"), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(path.Join(setupDir, "server", "setup.sh"))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, setupScript)
	if err != nil {
		return err
	}

	file, err = os.Create(path.Join(setupDir, "server", "start.sh"))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, startScript)
	if err != nil {
		return err
	}

	file, err = os.Create(path.Join(setupDir, "server", "configuration.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, configurationScript)
	if err != nil {
		return err
	}

	return nil
}

func (e OpenVSCodeServerPlugin) ProjectInit(project workspace.Project) error {
	execConfig := types.ExecConfig{
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd: []string{
			"bash",
			"/setup/server/setup.sh",
		},
		User: "daytona",
	}

	// TODO: Implement
	execResult, err := util.DockerExec("project.GetContainerName()", execConfig, nil)
	if err != nil {
		return err
	}

	if execResult.ExitCode != 0 {
		log.Error(execResult.StdErr)
		return errors.New("failed to initialize vscode server")
	}

	return nil
}

func (e OpenVSCodeServerPlugin) ProjectStart(project workspace.Project) error {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}

	execConfig := types.ExecConfig{
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd: []string{
			"bash",
			"/setup/server/start.sh",
		},
		User: "daytona",
	}

	// TODO: Implement
	execResp, err := cli.ContainerExecCreate(ctx, "project.GetContainerName()", execConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = cli.ContainerExecStart(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (e OpenVSCodeServerPlugin) ProjectLivenessProbe(project workspace.Project) (bool, error) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, err
	}

	// TODO: Implement
	inspect, err := cli.ContainerInspect(ctx, "project.GetContainerName()")
	if err != nil {
		return false, err
	}

	ip := inspect.NetworkSettings.Networks[project.Workspace.Name].IPAddress

	resp, err := http.Get(fmt.Sprintf("http://%s:%d", ip, IDE_PORT))
	if err != nil {
		log.WithFields(log.Fields{
			"project":   project.Name,
			"extension": e.GetName(),
		}).Debug(err)
		//	ignore err
		return false, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.WithFields(log.Fields{
			"project":   project.Name,
			"extension": e.GetName(),
		}).Debug(resp.StatusCode)
		return false, nil
	}

	return true, nil
}

func (e OpenVSCodeServerPlugin) ProjectLivenessProbeTimeout() int {
	return 30
}

func (e OpenVSCodeServerPlugin) ProjectInfo(project workspace.Project) string {
	//	todo: no tasilscale
	return ""
}
