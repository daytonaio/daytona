// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"

	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"
)

const UseUserHomeAsWorkDir = "DAYTONA_USER_HOME_AS_WORKDIR"

func (d *DockerClient) startDaytonaDaemon(ctx context.Context, containerId string, workdir string) error {
	defer timer.Timer()()

	daemonCmd := "/usr/local/bin/daytona"
	if workdir == "" {
		workdir = UseUserHomeAsWorkDir
	}
	daemonCmd = fmt.Sprintf("%s --workdir %s", daemonCmd, workdir)

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
