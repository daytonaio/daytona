// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"strings"

	"github.com/daytonaio/daytona/agent/workspace"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/plugin"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type SshPlugin struct {
	PublicKey string
	BasePath  string
}

var Config *plugin.WorkspacePluginConfig

func (e SshPlugin) SetConfig(config plugin.WorkspacePluginConfig) error {
	Config = &config
	return nil
}

func (e SshPlugin) GetVersion() string {
	return "0.0.1"
}

func (e SshPlugin) GetName() string {
	return "ssh"
}

func (e SshPlugin) ProjectPreInit(project workspace.Project) error {
	setupDir := Config.SetupPath

	err := os.MkdirAll(path.Join(setupDir, "ssh"), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(path.Join(setupDir, "ssh", "setup.sh"))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, strings.ReplaceAll(strings.ReplaceAll(setupScript, "{{user}}", "daytona"), "{{sshPublicKey}}", e.PublicKey))
	if err != nil {
		return err
	}

	file, err = os.Create(path.Join(setupDir, "ssh", "start.sh"))
	if err != nil {
		return err
	}
	defer file.Close()

	wsEnv := ""
	for _, envVar := range Config.EnvVars {
		wsEnv += envVar + " "
	}

	startScriptEnv := strings.Replace(startScript, "{{env}}", wsEnv, 1)

	_, err = io.WriteString(file, startScriptEnv)
	if err != nil {
		return err
	}

	file, err = os.Create(path.Join(setupDir, "ssh", "sshd_config"))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, strings.ReplaceAll(sshdConfig, "{{user}}", "daytona"))
	if err != nil {
		return err
	}

	return nil
}

func (e SshPlugin) ProjectInit(project workspace.Project) error {
	execConfig := types.ExecConfig{
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd: []string{
			"bash",
			"/setup/ssh/setup.sh",
		},
		User: "root",
	}

	// TODO: Implement
	execResult, err := util.DockerExec("project.GetContainerName()", execConfig, nil)
	if err != nil {
		return err
	}

	if execResult.ExitCode != 0 {
		log.Error(execResult.StdOut)
		log.Error(execResult.StdErr)
		return errors.New("failed to initialize openssh-server")
	}

	return nil
}

func (e SshPlugin) ProjectStart(project workspace.Project) error {
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
			"/setup/ssh/start.sh",
		},
		User: "root",
	}

	// TODO: Implement
	execResp, err := cli.ContainerExecCreate(ctx, "project.GetContainerName()", execConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = cli.ContainerExecStart(ctx, execResp.ID, types.ExecStartCheck{
		Detach: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (e SshPlugin) ProjectLivenessProbe(project workspace.Project) (bool, error) {
	return false, errors.New("not implemented")
	/*
		containerInfo, err := project.GetContainerInfo()
		if err != nil {
			return false, err
		}

		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:22", containerInfo.IP), 3*time.Second)
		if err != nil {
			log.WithFields(log.Fields{
				"project":   project.GetName(),
				"extension": e.Name(),
			}).Debug(err)
			return false, nil
		}
		defer conn.Close()

		return true, nil
	*/
}

func (e SshPlugin) ProjectLivenessProbeTimeout() int {
	return 60
}

func (e SshPlugin) ProjectInfo(project workspace.Project) string {
	//	todo: no tailscale
	return ""
}
