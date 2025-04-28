// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) startTerminalProcess(ctx context.Context, containerId string, port int) error {
	shell, err := d.detectShell(ctx, containerId)
	if err != nil {
		log.Errorf("Error detecting shell: %v, defaulting to sh", err)
		shell = "sh"
	}

	// Start terminal process
	terminalExecConfig := container.ExecOptions{
		Cmd:          []string{"terminal", "-p", strconv.Itoa(port), "-W", shell},
		AttachStdout: false,
		AttachStderr: false,
		Tty:          true,
	}

	terminalExecResp, err := d.apiClient.ContainerExecCreate(ctx, containerId, terminalExecConfig)
	if err != nil {
		log.Errorf("Error creating terminal process: %v", err)
		return nil
	}

	err = d.apiClient.ContainerExecStart(ctx, terminalExecResp.ID, container.ExecStartOptions{Detach: true})
	if err != nil {
		log.Errorf("Error starting terminal process: %v", err)
		return nil
	}

	return nil
}

func (d *DockerClient) detectShell(ctx context.Context, containerId string) (string, error) {
	execOptions := container.ExecOptions{
		Cmd:          []string{"which", "bash"},
		AttachStdout: true,
		AttachStderr: true,
	}

	result, err := d.execSync(ctx, containerId, execOptions, container.ExecStartOptions{})
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(string(result.StdOut)) != "" {
		return "bash", nil
	}

	return "sh", nil
}
