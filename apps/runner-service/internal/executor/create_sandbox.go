/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"encoding/json"
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
	e.log.Debug("creating sandbox", "job_id", job.GetId(), "sandbox_id", sandboxId)

	// Parse payload
	payload := job.GetPayload()
	var parsedPayload map[string]interface{}
	err := json.Unmarshal([]byte(payload), &parsedPayload)
	if err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	snapshot, _ := parsedPayload["snapshot"].(string)
	cpu, _ := parsedPayload["cpu"].(float64)
	mem, _ := parsedPayload["mem"].(float64)
	entrypoint, _ := parsedPayload["entrypoint"].([]interface{})

	// Convert entrypoint to string slice
	var entrypointStr []string
	for _, v := range entrypoint {
		if s, ok := v.(string); ok {
			entrypointStr = append(entrypointStr, s)
		}
	}

	// Get env map
	envMap := make(map[string]string)
	if envPayload, ok := parsedPayload["env"].(map[string]interface{}); ok {
		for k, v := range envPayload {
			if s, ok := v.(string); ok {
				envMap[k] = s
			}
		}
	}

	// Get labels map
	labelsMap := make(map[string]string)
	if labelsPayload, ok := parsedPayload["labels"].(map[string]interface{}); ok {
		for k, v := range labelsPayload {
			if s, ok := v.(string); ok {
				labelsMap[k] = s
			}
		}
	}

	// Get registry auth if present
	var authStr string
	if registryPayload, ok := parsedPayload["registry"].(map[string]interface{}); ok {
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
		e.log.Debug("pulling image", "image", snapshot)
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
		e.log.Debug("image exists", "image", snapshot)
	}

	// Inspect image to get working directory
	imageInspect, err := e.dockerClient.ImageInspect(ctx, snapshot)
	if err != nil {
		return fmt.Errorf("inspect image: %w", err)
	}

	workingDir := imageInspect.Config.WorkingDir

	// Add environment variable if no working dir
	if workingDir == "" {
		envMap["DAYTONA_USER_HOME_AS_WORKDIR"] = "true"
	}

	e.log.Debug("container config prepared",
		"entrypoint", DAEMON_PATH,
		"cmd", entrypointStr,
		"working_dir", workingDir)

	// Prepare container config with daemon as entrypoint
	// The payload entrypoint is passed as CMD to the daemon
	containerConfig := &container.Config{
		Image:      snapshot,
		Entrypoint: []string{DAEMON_PATH},
		Cmd:        entrypointStr,
		WorkingDir: workingDir,
		Env:        envMapToSlice(envMap),
		Labels:     labelsMap,
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

	e.log.Debug("container created", "container_id", resp.ID[:12])

	// Start container
	if err := e.dockerClient.ContainerStart(ctx, sandboxId, container.StartOptions{}); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	e.log.Debug("container started", "sandbox_id", sandboxId)

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

	// Daemon is the entrypoint, so it's already started with the container

	// Wait for daemon to be ready
	e.log.Debug("waiting for daemon to be ready", "sandbox_id", sandboxId, "ip", containerIP)
	if err := e.waitForDaemonRunning(ctx, containerIP, 10*time.Second); err != nil {
		e.log.Error("daemon failed to start", "error", err)
		return fmt.Errorf("daemon not ready: %w", err)
	}

	e.log.Debug("daemon is ready", "sandbox_id", sandboxId)

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
