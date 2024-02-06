// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package vsc_server

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

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type VscServerExtension struct {
}

func (e VscServerExtension) Name() string {
	return "vsc-server"
}

func (e VscServerExtension) PreInit(project workspace.Project) error {
	setupDir := project.GetSetupPath()

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

func (e VscServerExtension) Init(project workspace.Project) error {
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

	execResult, err := util.DockerExec(project.GetContainerName(), execConfig, nil)
	if err != nil {
		return err
	}

	if execResult.ExitCode != 0 {
		log.Error(execResult.StdErr)
		return errors.New("failed to initialize vscode server")
	}

	return nil
}

func (e VscServerExtension) Start(project workspace.Project) error {
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
	execResp, err := cli.ContainerExecCreate(ctx, project.GetContainerName(), execConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = cli.ContainerExecStart(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (e VscServerExtension) LivenessProbe(project workspace.Project) (bool, error) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, err
	}

	inspect, err := cli.ContainerInspect(ctx, project.GetContainerName())
	if err != nil {
		return false, err
	}

	ip := inspect.NetworkSettings.Networks[project.Workspace.Name].IPAddress

	resp, err := http.Get(fmt.Sprintf("http://%s:%d", ip, IDE_PORT))
	if err != nil {
		log.WithFields(log.Fields{
			"project":   project.GetName(),
			"extension": e.Name(),
		}).Debug(err)
		//	ignore err
		return false, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.WithFields(log.Fields{
			"project":   project.GetName(),
			"extension": e.Name(),
		}).Debug(resp.StatusCode)
		return false, nil
	}

	return true, nil
}

func (e VscServerExtension) LivenessProbeTimeout() int {
	return 30
}

func (e VscServerExtension) Info(project workspace.Project) string {
	//	todo: no tasilscale
	return ""
}
