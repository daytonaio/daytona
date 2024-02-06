// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/daytonaio/daytona/agent/workspace"
	"github.com/daytonaio/daytona/internal/util"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type SshExtension struct {
	PublicKey string
}

func (e SshExtension) Name() string {
	return "ssh"
}

func (e SshExtension) PreInit(project workspace.Project) error {
	setupDir := project.GetSetupPath()

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
	for _, envVar := range project.GetEnvVars() {
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

func (e SshExtension) Init(project workspace.Project) error {
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

	execResult, err := util.DockerExec(project.GetContainerName(), execConfig, nil)
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

func (e SshExtension) Start(project workspace.Project) error {
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
	execResp, err := cli.ContainerExecCreate(ctx, project.GetContainerName(), execConfig)
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

func (e SshExtension) LivenessProbe(project workspace.Project) (bool, error) {
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
}

func (e SshExtension) LivenessProbeTimeout() int {
	return 60
}

func (e SshExtension) Info(project workspace.Project) string {
	//	todo: no tailscale
	return ""
}
