// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	log "github.com/sirupsen/logrus"
)

const (
	ToolboxContainerName  = "mock-runner-toolbox"
	ToolboxImage          = "ubuntu:22.04"
	ToolboxDaemonPort     = 2280
	DaemonStartTimeoutSec = 60
)

// ToolboxContainer manages the shared container used for all toolbox operations
type ToolboxContainer struct {
	client       client.APIClient
	containerID  string
	containerIP  string
	daemonPath   string
	mu           sync.RWMutex
	isRunning    bool
}

// NewToolboxContainer creates a new toolbox container manager
func NewToolboxContainer(cli client.APIClient, daemonPath string) *ToolboxContainer {
	return &ToolboxContainer{
		client:     cli,
		daemonPath: daemonPath,
	}
}

// Start ensures the toolbox container is running with the daemon
func (tc *ToolboxContainer) Start(ctx context.Context) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Check if container already exists
	containers, err := tc.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if name == "/"+ToolboxContainerName {
				tc.containerID = c.ID
				if c.State == "running" {
					tc.isRunning = true
					// Get container IP
					inspect, err := tc.client.ContainerInspect(ctx, c.ID)
					if err != nil {
						return fmt.Errorf("failed to inspect existing container: %w", err)
					}
					tc.containerIP = tc.getContainerIP(inspect)
					log.Infof("Toolbox container already running with IP: %s", tc.containerIP)
					
					// Ensure daemon is running
					return tc.ensureDaemonRunning(ctx)
				}
				// Container exists but not running, start it
				return tc.startExistingContainer(ctx)
			}
		}
	}

	// Container doesn't exist, create it
	return tc.createAndStartContainer(ctx)
}

// getContainerIP extracts the IP address from container inspect response
func (tc *ToolboxContainer) getContainerIP(inspect container.InspectResponse) string {
	for _, network := range inspect.NetworkSettings.Networks {
		if network.IPAddress != "" {
			return network.IPAddress
		}
	}
	return ""
}

// createAndStartContainer creates a new toolbox container and starts it
func (tc *ToolboxContainer) createAndStartContainer(ctx context.Context) error {
	log.Info("Creating toolbox container...")

	// Pull the image first
	reader, err := tc.client.ImagePull(ctx, ToolboxImage, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", ToolboxImage, err)
	}
	defer reader.Close()
	io.Copy(io.Discard, reader) // Wait for pull to complete

	// Create the container with a long-running command
	containerConfig := &container.Config{
		Image:      ToolboxImage,
		Cmd:        []string{"sleep", "infinity"},
		WorkingDir: "/home/daytona",
		Tty:        true,
	}

	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyUnlessStopped,
		},
	}

	networkConfig := &network.NetworkingConfig{}

	resp, err := tc.client.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, &v1.Platform{
		Architecture: "amd64",
		OS:           "linux",
	}, ToolboxContainerName)
	if err != nil {
		return fmt.Errorf("failed to create toolbox container: %w", err)
	}

	tc.containerID = resp.ID
	log.Infof("Created toolbox container with ID: %s", tc.containerID)

	return tc.startExistingContainer(ctx)
}

// startExistingContainer starts an existing container and the daemon
func (tc *ToolboxContainer) startExistingContainer(ctx context.Context) error {
	log.Infof("Starting toolbox container %s...", tc.containerID)

	err := tc.client.ContainerStart(ctx, tc.containerID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start toolbox container: %w", err)
	}

	// Wait for container to be running
	for i := 0; i < 30; i++ {
		inspect, err := tc.client.ContainerInspect(ctx, tc.containerID)
		if err != nil {
			return fmt.Errorf("failed to inspect container: %w", err)
		}
		if inspect.State.Running {
			tc.containerIP = tc.getContainerIP(inspect)
			tc.isRunning = true
			log.Infof("Toolbox container started with IP: %s", tc.containerIP)
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !tc.isRunning {
		return fmt.Errorf("toolbox container failed to start")
	}

	// Copy daemon binary to container and start it
	return tc.startDaemon(ctx)
}

// startDaemon copies the daemon binary into the container and starts it
func (tc *ToolboxContainer) startDaemon(ctx context.Context) error {
	if tc.daemonPath == "" {
		log.Warn("No daemon path configured, skipping daemon start")
		return nil
	}

	// Check if daemon binary exists
	if _, err := os.Stat(tc.daemonPath); os.IsNotExist(err) {
		log.Warnf("Daemon binary not found at %s, skipping daemon start", tc.daemonPath)
		return nil
	}

	log.Info("Copying daemon binary to toolbox container...")

	// Create a tar archive of the daemon binary for copying
	tarReader, err := createTarFromFile(tc.daemonPath, "daemon")
	if err != nil {
		return fmt.Errorf("failed to create tar archive: %w", err)
	}

	// Copy to container
	err = tc.client.CopyToContainer(ctx, tc.containerID, "/tmp", tarReader, container.CopyToContainerOptions{})
	if err != nil {
		return fmt.Errorf("failed to copy daemon to container: %w", err)
	}

	// Make daemon executable and start it in background
	execConfig := container.ExecOptions{
		Cmd:          []string{"sh", "-c", "chmod +x /tmp/daemon && nohup /tmp/daemon > /tmp/daemon.log 2>&1 &"},
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := tc.client.ContainerExecCreate(ctx, tc.containerID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec for daemon start: %w", err)
	}

	err = tc.client.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	log.Info("Daemon started in toolbox container")

	// Wait for daemon to be ready
	return tc.waitForDaemon(ctx)
}

// ensureDaemonRunning checks if daemon is running and starts it if not
func (tc *ToolboxContainer) ensureDaemonRunning(ctx context.Context) error {
	// Try to connect to daemon
	targetURL := fmt.Sprintf("http://%s:%d/version", tc.containerIP, ToolboxDaemonPort)
	target, err := url.Parse(targetURL)
	if err != nil {
		return err
	}

	conn, err := net.DialTimeout("tcp", target.Host, 2*time.Second)
	if err != nil {
		// Daemon not running, start it
		log.Info("Daemon not responding, starting it...")
		return tc.startDaemon(ctx)
	}
	conn.Close()
	log.Info("Daemon is running")
	return nil
}

// waitForDaemon waits for the daemon to become ready
func (tc *ToolboxContainer) waitForDaemon(ctx context.Context) error {
	targetURL := fmt.Sprintf("http://%s:%d/version", tc.containerIP, ToolboxDaemonPort)
	target, err := url.Parse(targetURL)
	if err != nil {
		return err
	}

	timeout := time.Duration(DaemonStartTimeoutSec) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	log.Infof("Waiting for daemon at %s...", target.Host)

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for daemon to start")
		default:
			conn, err := net.DialTimeout("tcp", target.Host, 1*time.Second)
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			conn.Close()
			log.Info("Daemon is ready")
			return nil
		}
	}
}

// GetIP returns the IP address of the toolbox container
func (tc *ToolboxContainer) GetIP() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.containerIP
}

// GetContainerID returns the container ID
func (tc *ToolboxContainer) GetContainerID() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.containerID
}

// IsRunning returns whether the container is running
func (tc *ToolboxContainer) IsRunning() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.isRunning
}

// Stop stops the toolbox container
func (tc *ToolboxContainer) Stop(ctx context.Context) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.containerID == "" {
		return nil
	}

	log.Info("Stopping toolbox container...")
	timeout := 10
	err := tc.client.ContainerStop(ctx, tc.containerID, container.StopOptions{Timeout: &timeout})
	if err != nil {
		return fmt.Errorf("failed to stop toolbox container: %w", err)
	}

	tc.isRunning = false
	log.Info("Toolbox container stopped")
	return nil
}



