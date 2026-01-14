// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"
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

func (d *DockerClient) waitForDaemonRunning(ctx context.Context, containerId, containerIP string, onCreateStart bool) error {
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
			ct, err := d.ContainerInspect(ctx, containerId)
			if err != nil {
				log.Errorf("Failed to inspect container on waiting for daemon to start error %s: %v", containerId, err)
				return fmt.Errorf("timeout waiting for daemon to start")
			}

			if ct.State != nil && !ct.State.Running {
				logs, err := d.readLastErrorLog(ctx, containerId)
				if err != nil {
					log.Errorf("Failed to read error log for container %s: %v", containerId, err)
					return fmt.Errorf("timeout waiting for daemon to start")
				}

				if logs != "" {
					logErr := fmt.Errorf("sandbox exited with error: %s", logs)
					if onCreateStart {
						return common_errors.NewBadRequestError(logErr)
					}
					return logErr
				}
			}

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

func (d *DockerClient) readLastErrorLog(ctx context.Context, containerId string) (string, error) {
	logs, err := d.apiClient.ContainerLogs(ctx, containerId, container.LogsOptions{
		ShowStdout: false,
		ShowStderr: true,
		Follow:     false,
		Tail:       "50", // Get more lines to ensure we find the last error
	})
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logs.Close()

	var logOutput []byte
	buf := make([]byte, 8192)
	for {
		n, err := logs.Read(buf)
		if n > 0 {
			logOutput = append(logOutput, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	// Sanitize logs and extract only the last error
	sanitizedLogs := sanitizeLogPrefixes(string(logOutput))

	return sanitizedLogs, nil
}

func sanitizeLogPrefixes(logs string) string {
	// Remove Docker's 8-byte log stream header
	// Format: [stream_type(1 byte)][padding(3 bytes)][size(4 bytes)]
	// Example: \u0002\u0000\u0000\u0000\u0000\u0000\u0000k or \u0002\u0000\u0000\u0000\u0000\u0000\u0000\
	dockerHeaderPattern := regexp.MustCompile(`[\x00-\x02][\x00]{3}[\x00-\xff]{4}`)
	sanitized := dockerHeaderPattern.ReplaceAllString(logs, "")

	// Remove ANSI color codes (e.g., \x1b[31m for red, \x1b[33m for yellow, etc.)
	ansiPattern := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	sanitized = ansiPattern.ReplaceAllString(sanitized, "")

	// Extract only lines starting with ERRO[0...] (e.g., ERRO[0000], ERRO[0010])
	errorPattern := regexp.MustCompile(`(?m)^ERRO\[\d+\].*$`)
	errorLines := errorPattern.FindAllString(sanitized, -1)

	// If no error lines found, return original sanitized content
	if len(errorLines) == 0 {
		// return empty string if no error found, to avoid leaking non-error logs
		return ""
	}

	// Take only the last error line
	lastErrorLine := errorLines[len(errorLines)-1]

	// Remove the ERRO[0...] prefix from the line
	erroPrefixPattern := regexp.MustCompile(`^ERRO\[\d+\]\s*`)
	sanitized = erroPrefixPattern.ReplaceAllString(lastErrorLine, "")

	return sanitized
}
