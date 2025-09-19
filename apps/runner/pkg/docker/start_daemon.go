// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"

	common_daemon "github.com/daytonaio/common-go/pkg/daemon"
	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) startDaytonaDaemon(ctx context.Context, containerId string, workDir string) error {
	defer timer.Timer()()

	daemonCmd := "/usr/local/bin/daytona"
	if workDir == "" {
		workDir = common_daemon.UseUserHomeAsWorkDir
	}
	daemonCmd = fmt.Sprintf("%s --work-dir %s", daemonCmd, workDir)

	execOptions := container.ExecOptions{
		Cmd:          []string{"sh", "-c", daemonCmd},
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}

	execStartOptions := container.ExecStartOptions{
		Detach: false,
	}

	result, err := d.execSync(ctx, containerId, execOptions, execStartOptions)
	if err != nil {
		log.Errorf("Error starting Daytona daemon: %s", err.Error())
		return nil
	}

	if result.ExitCode != 0 && result.StdErr != "" {
		log.Errorf("Error starting Daytona daemon: %s", string(result.StdErr))
		return nil
	}

	return nil
}
