// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	common_daemon "github.com/daytonaio/common-go/pkg/daemon"
	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) startDaytonaDaemon(ctx context.Context, containerId string, workDir string) error {
	defer timer.Timer()()

	envVars := []string{}
	if workDir == "" {
		envVars = append(envVars, fmt.Sprintf("%s=true", common_daemon.UserHomeAsWorkDirEnvVar))
	}

	execOptions := container.ExecOptions{
		Cmd:          []string{"/usr/local/bin/daytona"},
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

	if result.ExitCode != 0 && result.StdErr != "" {
		return fmt.Errorf("failed to start daytona daemon: %s", result.StdErr)
	}

	return nil
}

func (d *DockerClient) waitForDaemonRunning(ctx context.Context, containerIP string, timeout time.Duration) error {
	defer timer.Timer()()

	// Build the target URL
	targetURL := fmt.Sprintf("http://%s:2280/version", containerIP)
	target, err := url.Parse(targetURL)
	if err != nil {
		return common_errors.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err))
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for daemon to start")
		default:
			conn, err := net.DialTimeout("tcp", target.Host, 1*time.Second)
			if err != nil {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			conn.Close()
			return nil
		}
	}
}
