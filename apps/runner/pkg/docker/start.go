// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

func (d *DockerClient) Start(ctx context.Context, containerId string, authToken *string, metadata map[string]string) error {
	defer timer.Timer()()
	d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStarting)

	// Cancel a backup if it's already in progress
	backup_context, ok := backup_context_map.Get(containerId)
	if ok {
		backup_context.cancel()
	}

	c, err := d.ContainerInspect(ctx, containerId)
	if err != nil {
		return err
	}

	if c.State.Running {
		containerIP := common.GetContainerIpAddress(ctx, c)
		if containerIP == "" {
			return errors.New("sandbox IP not found? Is the sandbox started?")
		}

		err = d.waitForDaemonRunning(ctx, containerIP, authToken)
		if err != nil {
			return err
		}

		d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStarted)
		return nil
	}

	err = d.apiClient.ContainerStart(ctx, containerId, container.StartOptions{})
	if err != nil {
		return err
	}

	// make sure container is running
	err = d.waitForContainerRunning(ctx, containerId)
	if err != nil {
		return err
	}

	c, err = d.ContainerInspect(ctx, containerId)
	if err != nil {
		return err
	}

	containerIP := common.GetContainerIpAddress(ctx, c)
	if containerIP == "" {
		return errors.New("sandbox IP not found? Is the sandbox started?")
	}

	processesCtx := context.Background()
	go func() {
		if err := d.startDaytonaDaemon(processesCtx, containerId, c.Config.WorkingDir); err != nil {
			log.Errorf("Failed to start Daytona daemon: %s\n", err.Error())
		}
	}()

	err = d.waitForDaemonRunning(ctx, containerIP, authToken)
	if err != nil {
		return err
	}

	d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStarted)

	if metadata["limitNetworkEgress"] == "true" {
		go func() {
			containerShortId := c.ID[:12]
			err = d.netRulesManager.SetNetworkLimiter(containerShortId, containerIP)
			if err != nil {
				log.Errorf("Failed to set network limiter: %v", err)
			}
		}()
	}

	return nil
}

func (d *DockerClient) waitForContainerRunning(ctx context.Context, containerId string) error {
	defer timer.Timer()()

	timeout := time.Duration(d.sandboxStartTimeoutSec) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return errors.New("timeout waiting for the sandbox to start - please ensure that your entrypoint is long-running")
		case <-ticker.C:
			c, err := d.ContainerInspect(ctx, containerId)
			if err != nil {
				return err
			}

			if c.State.Running {
				return nil
			}
		}
	}
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
