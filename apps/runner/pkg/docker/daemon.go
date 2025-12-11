// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"net/url"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) startDaytonaDaemon(ctx context.Context, containerId string, workDir string) error {
	defer timer.Timer()()

	var envVars []string
	if workDir == "" {
		envVars = append(envVars, "DAYTONA_USER_HOME_AS_WORKDIR=true")
	}

	execOptions := container.ExecOptions{
		Cmd:          []string{common.DAEMON_PATH},
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   workDir,
		Env:          envVars,
		Tty:          true,
	}

	execStartOptions := container.ExecStartOptions{
		Detach: false,
	}

	result, err := d.execSync(ctx, containerId, execOptions, execStartOptions)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("failed to start daytona daemon with exit code %d: %s", result.ExitCode, result.StdErr)
	}

	return nil
}

func (d *DockerClient) waitForDaemonRunning(ctx context.Context, containerIP string) (string, error) {
	defer timer.Timer()()

	// Build the target URL
	targetURL := fmt.Sprintf("http://%s:2280/version", containerIP)
	target, err := url.Parse(targetURL)
	if err != nil {
		return "", common_errors.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err))
	}

	timeout := time.Duration(d.daemonStartTimeoutSec) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return "", fmt.Errorf("timeout waiting for daemon to start")
		default:
			version, err := d.getDaemonVersion(ctx, target)
			if err != nil {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			return version, nil
		}
	}
}
