/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	apiclient "github.com/daytonaio/apiclient"
)

func (e *Executor) createSandbox(ctx context.Context, job *apiclient.Job) error {
	sandboxId := job.GetResourceId()
	e.log.Info("creating sandbox", "job_id", job.GetId(), "sandbox_id", sandboxId)

	// Parse payload
	payload := job.GetPayload()

	snapshot, _ := payload["snapshot"].(string)
	cpu, _ := payload["cpu"].(float64)
	mem, _ := payload["mem"].(float64)
	entrypoint, _ := payload["entrypoint"].([]interface{})

	// Convert entrypoint to string slice
	var entrypointStr []string
	for _, v := range entrypoint {
		if s, ok := v.(string); ok {
			entrypointStr = append(entrypointStr, s)
		}
	}

	// Get env map
	envMap := make(map[string]string)
	if envPayload, ok := payload["env"].(map[string]interface{}); ok {
		for k, v := range envPayload {
			if s, ok := v.(string); ok {
				envMap[k] = s
			}
		}
	}

	// Get labels map
	labelsMap := make(map[string]string)
	if labelsPayload, ok := payload["labels"].(map[string]interface{}); ok {
		for k, v := range labelsPayload {
			if s, ok := v.(string); ok {
				labelsMap[k] = s
			}
		}
	}

	// Get registry auth if present
	var authStr string
	if registryPayload, ok := payload["registry"].(map[string]interface{}); ok {
		username, _ := registryPayload["username"].(string)
		password, _ := registryPayload["password"].(string)
		server, _ := registryPayload["server"].(string)

		authConfig := registry.AuthConfig{
			Username:      username,
			Password:      password,
			ServerAddress: server,
		}
		if encoded, err := registry.EncodeAuthConfig(authConfig); err == nil {
			authStr = encoded
		}
	}

	// Check if image exists
	imageExists, err := e.dockerClient.ImageList(ctx, image.ListOptions{Filters: filters.NewArgs(filters.Arg("reference", snapshot))})
	if err != nil {
		return fmt.Errorf("check image exists: %w", err)
	}
	if len(imageExists) == 0 {
		// Pull image
		e.log.Info("pulling image", "image", snapshot)
		reader, err := e.dockerClient.ImagePull(ctx, snapshot, image.PullOptions{RegistryAuth: authStr})
		if err != nil {
			return fmt.Errorf("pull image: %w", err)
		}
		// Consume output
		buf := make([]byte, 4096)
		for {
			if _, err := reader.Read(buf); err != nil {
				break
			}
		}
		reader.Close()
	} else {
		e.log.Info("image exists", "image", snapshot)
	}

	// Prepare container config
	containerConfig := &container.Config{
		Image:      snapshot,
		Entrypoint: entrypointStr,
		Env:        envMapToSlice(envMap),
		Labels:     labelsMap,
		// User and WorkingDir are not set initially - the daemon/entrypoint will create the user
		// WorkingDir: fmt.Sprintf("/home/%s", osUser),
		// User:       osUser,
	}

	// Prepare binds - mount daemon binary
	binds := []string{
		fmt.Sprintf("%s:/usr/local/bin/daytona:ro", e.daemonPath),
	}

	// Prepare host config
	hostConfig := &container.HostConfig{
		NetworkMode: "bridge",
		Binds:       binds,
		Resources: container.Resources{
			NanoCPUs: int64(cpu) * 1e9,
			Memory:   int64(mem) * 1024 * 1024 * 1024,
		},
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
	}

	// Create container
	resp, err := e.dockerClient.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		nil,
		&v1.Platform{Architecture: "amd64", OS: "linux"},
		sandboxId,
	)
	if err != nil {
		return fmt.Errorf("create container: %w", err)
	}

	e.log.Info("container created", "container_id", resp.ID[:12])

	// Start container
	if err := e.dockerClient.ContainerStart(ctx, sandboxId, container.StartOptions{}); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	e.log.Info("container started", "sandbox_id", sandboxId)

	// Start Daytona daemon inside the container
	daemonCmd := "/usr/local/bin/daytona"
	execConfig := container.ExecOptions{
		Cmd:          []string{"sh", "-c", daemonCmd},
		AttachStdout: true,
		AttachStderr: true,
		Detach:       true, // Run in background
	}

	execResp, err := e.dockerClient.ContainerExecCreate(ctx, sandboxId, execConfig)
	if err != nil {
		e.log.Error("failed to create daemon exec", "error", err)
		return fmt.Errorf("create daemon exec: %w", err)
	}

	if err := e.dockerClient.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{Detach: true}); err != nil {
		e.log.Error("failed to start daemon", "error", err)
		return fmt.Errorf("start daemon: %w", err)
	}

	e.log.Info("daemon exec started", "sandbox_id", sandboxId)

	// Get container IP for daemon health check
	containerInfo, err := e.dockerClient.ContainerInspect(ctx, sandboxId)
	if err != nil {
		e.log.Error("failed to inspect container", "error", err)
		return fmt.Errorf("inspect container: %w", err)
	}

	containerIP := ""
	for _, network := range containerInfo.NetworkSettings.Networks {
		containerIP = network.IPAddress
		break
	}

	if containerIP == "" {
		return fmt.Errorf("no IP address found for container")
	}

	// Wait for daemon to be ready
	e.log.Info("waiting for daemon to be ready", "sandbox_id", sandboxId, "ip", containerIP)
	if err := e.waitForDaemonRunning(ctx, containerIP, 10*time.Second); err != nil {
		e.log.Error("daemon failed to start", "error", err)
		return fmt.Errorf("daemon not ready: %w", err)
	}

	e.log.Info("daemon is ready", "sandbox_id", sandboxId)

	// Update allocations
	e.collector.IncrementAllocations(float32(cpu), float32(mem), 0)

	e.log.Info("sandbox created successfully", "sandbox_id", sandboxId)
	return nil
}

// waitForDaemonRunning waits for the Daytona daemon to start accepting connections on port 2280
func (e *Executor) waitForDaemonRunning(ctx context.Context, containerIP string, timeout time.Duration) error {
	// Create a span for daemon health check
	tracer := otel.Tracer("runner-service")
	ctx, span := tracer.Start(ctx, "wait_for_daemon_ready") // Use WithAttributes directly in Start options

	span.SetAttributes(
		attribute.String("daemon.address", containerIP),
		attribute.String("daemon.port", "2280"),
		attribute.Float64("timeout.seconds", timeout.Seconds()),
	)
	defer span.End()

	daemonAddr := fmt.Sprintf("%s:2280", containerIP)

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	attemptCount := 0
	for {
		select {
		case <-timeoutCtx.Done():
			span.SetAttributes(
				attribute.Int("attempts", attemptCount),
				attribute.Bool("success", false),
			)
			err := fmt.Errorf("timeout waiting for daemon to start on %s", daemonAddr)
			span.RecordError(err)
			return err
		case <-ticker.C:
			attemptCount++
			conn, err := net.DialTimeout("tcp", daemonAddr, 1*time.Second)
			if err == nil {
				conn.Close()
				span.SetAttributes(
					attribute.Int("attempts", attemptCount),
					attribute.Bool("success", true),
				)
				return nil
			}
			// Continue waiting
		}
	}
}
