// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
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

func (d *DockerClient) waitForDaemonRunning(ctx context.Context, containerIP string, authToken *string) error {
	defer timer.Timer()()

	// Build the target URL
	targetURL := fmt.Sprintf("http://%s:2280/version", containerIP)
	target, err := url.Parse(targetURL)
	if err != nil {
		return common_errors.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err))
	}

	timeout := time.Duration(d.daemonStartTimeoutSec) * time.Second
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

			// For backward compatibility, only initialize daemon if authToken is provided
			if authToken == nil {
				return nil
			}

			return d.initializeDaemon(containerIP, *authToken)
		}
	}
}

type sandboxToken struct {
	Token string `json:"token"`
}

func (d *DockerClient) initializeDaemon(containerIP string, token string) error {
	sandboxToken := sandboxToken{
		Token: token,
	}

	jsonData, err := json.Marshal(sandboxToken)
	if err != nil {
		return fmt.Errorf("failed to marshal sandbox token data: %w", err)
	}

	url := fmt.Sprintf("http://%s:2280/init", containerIP)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to initialize daemon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("daemon returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}