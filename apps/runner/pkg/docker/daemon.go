// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/docker/docker/api/types/container"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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

func (d *DockerClient) waitForDaemonRunning(ctx context.Context, containerIP string, authToken *string) (string, error) {
	defer timer.Timer()()

	tracer := otel.Tracer("runner")
	ctx, span := tracer.Start(ctx, "wait_for_daemon_running",
		trace.WithAttributes(attribute.String("container.ip", containerIP)),
	)
	defer span.End()

	// Build the target URL
	targetURL := fmt.Sprintf("http://%s:2280/version", containerIP)
	target, err := url.Parse(targetURL)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to parse target URL")
		return "", common_errors.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err))
	}

	// Use a plain HTTP client for polling to avoid creating noisy trace spans for each failed retry
	pollingClient := &http.Client{
		Timeout: 1 * time.Second,
	}

	timeout := time.Duration(d.daemonStartTimeoutSec) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	retries := 0
	for {
		select {
		case <-timeoutCtx.Done():
			span.SetAttributes(attribute.Int("retries", retries))
			span.RecordError(fmt.Errorf("timeout waiting for daemon to start"))
			span.SetStatus(codes.Error, "timeout waiting for daemon to start")
			return "", fmt.Errorf("timeout waiting for daemon to start")
		default:
			version, err := d.getDaemonVersion(ctx, target, pollingClient)
			if err != nil {
				retries++
				time.Sleep(5 * time.Millisecond)
				continue
			}

			span.SetAttributes(attribute.Int("retries", retries))
			span.SetStatus(codes.Ok, "daemon ready")

			// For backward compatibility, only initialize daemon if authToken is provided
			if authToken == nil {
				return version, nil
			}

			// Optimistically initialize the daemon in parallel while waiting for it to be ready, to save time.
			// If initialization fails, log the error but do not fail the entire process, as the daemon is already running at this point.
			otelClient := &http.Client{
				Timeout:   1 * time.Second,
				Transport: otelhttp.NewTransport(http.DefaultTransport),
			}
			go func() {
				// Don't cancel context
				initContext := context.WithoutCancel(ctx)
				err := d.initializeDaemon(initContext, containerIP, *authToken, otelClient)
				if err != nil {
					d.logger.ErrorContext(initContext, "Failed to initialize daemon telemetry", "error", err)
				}
			}()

			return version, nil
		}
	}
}

type sandboxToken struct {
	Token string `json:"token"`
}

func (d *DockerClient) initializeDaemon(ctx context.Context, containerIP string, token string, client *http.Client) error {
	if client == nil {
		return fmt.Errorf("http client is nil")
	}

	if !d.initializeDaemonTelemetry {
		return nil
	}

	sandboxToken := sandboxToken{
		Token: token,
	}

	jsonData, err := json.Marshal(sandboxToken)
	if err != nil {
		return fmt.Errorf("failed to marshal sandbox token data: %w", err)
	}

	url := fmt.Sprintf("http://%s:2280/init", containerIP)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create init request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to initialize daemon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("daemon returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}
