// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type FeatureItem struct {
	ID         string `json:"id"`
	Entrypoint string `json:"entrypoint,omitempty"`
	// Add other fields as needed
}

func (d *DockerClient) startImageWorkspace(opts *CreateWorkspaceOptions) error {
	containerName := d.GetWorkspaceContainerName(opts.Workspace)
	ctx := context.Background()

	c, err := d.apiClient.ContainerInspect(ctx, containerName)
	if err != nil {
		return err
	}

	// TODO: Add logging
	_, composeContainers, err := d.getComposeContainers(c)
	if err != nil {
		return err
	}

	if composeContainers != nil {
		if opts.LogWriter != nil {
			opts.LogWriter.Write([]byte("Starting compose containers\n"))
		}

		for _, c := range composeContainers {
			err = d.apiClient.ContainerStart(ctx, c.ID, container.StartOptions{})
			if err != nil {
				return err
			}
			if opts.LogWriter != nil {
				opts.LogWriter.Write([]byte(fmt.Sprintf("Started %s\n", strings.TrimPrefix(c.Names[0], "/"))))
			}
		}
	}

	if err == nil && c.State.Running {
		return nil
	}

	err = d.apiClient.ContainerStart(ctx, containerName, container.StartOptions{})
	if err != nil {
		return err
	}

	// make sure container is running
	//	TODO: timeout
	for {
		c, err = d.apiClient.ContainerInspect(ctx, containerName)
		if err != nil {
			return err
		}

		if c.State.Running {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	//	Find entrypoint metadata
	//	These entrypoints are used to run commands after the container is started (e.g. dockerd)
	c, err = d.apiClient.ContainerInspect(ctx, containerName)
	if err != nil {
		return err
	}

	//	First check if the image was built using devcontainer

	// Check if the "devcontainer.metadata" label exists
	metadata, ok := c.Config.Labels["devcontainer.metadata"]
	if ok {
		opts.LogWriter.Write([]byte("Found devcontainer.metadata label\n"))
		// Parse the metadata JSON
		var features []FeatureItem
		err = json.Unmarshal([]byte(metadata), &features)
		if err != nil {
			opts.LogWriter.Write([]byte(fmt.Sprintf("Failed to parse devcontainer.metadata: %v", err)))
			return nil
		}

		// Execute entrypoints
		err = executeEntrypoints(ctx, d.apiClient, c.ID, features, opts)
		if err != nil {
			opts.LogWriter.Write([]byte(fmt.Sprintf("Failed to execute entrypoints: %v", err)))
			return nil
		}
	}

	//	TODO: add daytona metadata support for images that are not based on the devcontainer config

	return nil
}

func executeEntrypoints(ctx context.Context, cli client.APIClient, containerID string, features []FeatureItem, opts *CreateWorkspaceOptions) error {
	for _, feature := range features {
		if feature.Entrypoint != "" {
			execConfig := container.ExecOptions{
				Cmd:          []string{"/bin/sh", "-c", feature.Entrypoint},
				AttachStdout: true,
				AttachStderr: true,
			}

			// Create the exec instance
			execIDResp, err := cli.ContainerExecCreate(ctx, containerID, execConfig)
			if err != nil {
				return fmt.Errorf("failed to create exec for feature %s: %v", feature.ID, err)
			}

			// Start the exec instance
			err = cli.ContainerExecStart(ctx, execIDResp.ID, container.ExecStartOptions{})
			if err != nil {
				return fmt.Errorf("failed to start exec for feature %s: %v", feature.ID, err)
			}

			opts.LogWriter.Write([]byte(fmt.Sprintf("Executed entrypoint for feature %s\n", feature.ID)))
		}
	}
	return nil
}
