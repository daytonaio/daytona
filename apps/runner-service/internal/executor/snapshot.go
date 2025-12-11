/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	apiclient "github.com/daytonaio/apiclient"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
)

func (e *Executor) buildSnapshot(ctx context.Context, job *apiclient.Job) error {
	// Get snapshot ref from resourceId
	snapshotRef := job.GetResourceId()
	if snapshotRef == "" {
		return fmt.Errorf("snapshotRef (resourceId) not found in job")
	}

	payload := job.GetPayload()

	e.log.Info("Building snapshot", slog.String("snapshot_ref", snapshotRef))

	// Extract buildInfo from payload
	buildInfo, ok := payload["buildInfo"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("buildInfo not found in payload")
	}

	e.log.Debug("Build info",
		slog.Any("build_info", buildInfo),
		slog.String("snapshot_ref", snapshotRef))

	// TODO: Implement actual snapshot build
	// - Build container image from Dockerfile or spec using buildInfo
	// - Tag with snapshot ref
	// - Push to internal registry

	// For now, return an error indicating not implemented
	// This prevents silent failures
	return fmt.Errorf("BUILD_SNAPSHOT not yet implemented for v2 runners")
}

func (e *Executor) pullSnapshot(ctx context.Context, job *apiclient.Job) error {
	// Get destination ref from resourceId (internal registry reference)
	destinationRef := job.GetResourceId()
	if destinationRef == "" {
		return fmt.Errorf("destinationRef (resourceId) not found in job")
	}

	payload := job.GetPayload()

	// Get source image from payload (external image like ubuntu:22.04)
	sourceImage, ok := payload["sourceImage"].(string)
	if !ok || sourceImage == "" {
		// Fallback to destination ref for backwards compatibility
		sourceImage = destinationRef
	}

	e.log.Info("Pulling snapshot",
		slog.String("source_image", sourceImage),
		slog.String("destination_ref", destinationRef))

	// Build image pull options with source registry auth
	pullOptions := image.PullOptions{
		Platform: "linux/amd64",
	}

	// Extract source registry credentials if present
	if registryPayload, ok := payload["registry"].(map[string]interface{}); ok {
		authConfig := registry.AuthConfig{}

		if username, ok := registryPayload["username"].(string); ok {
			authConfig.Username = username
		}
		if password, ok := registryPayload["password"].(string); ok {
			authConfig.Password = password
		}

		// Encode auth config to base64
		if authConfig.Username != "" || authConfig.Password != "" {
			encodedJSON, err := json.Marshal(authConfig)
			if err != nil {
				return fmt.Errorf("failed to encode registry auth: %w", err)
			}
			pullOptions.RegistryAuth = base64.URLEncoding.EncodeToString(encodedJSON)
		}
	}

	// Pull the source image
	reader, err := e.dockerClient.ImagePull(ctx, sourceImage, pullOptions)
	if err != nil {
		return fmt.Errorf("failed to pull snapshot: %w", err)
	}
	defer reader.Close()

	// Consume the output to wait for pull to complete
	buf := make([]byte, 4096)
	for {
		_, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			break
		}
	}

	// If destination is different from source, tag and push to internal registry
	if destinationRef != sourceImage {
		e.log.Debug("Tagging image for internal registry",
			slog.String("source", sourceImage),
			slog.String("destination", destinationRef))

		// Tag the image
		if err := e.dockerClient.ImageTag(ctx, sourceImage, destinationRef); err != nil {
			return fmt.Errorf("failed to tag snapshot: %w", err)
		}

		// Push to internal registry if destination registry credentials are provided
		if destRegistryPayload, ok := payload["destinationRegistry"].(map[string]interface{}); ok {
			pushOptions := image.PushOptions{}

			authConfig := registry.AuthConfig{}
			if username, ok := destRegistryPayload["username"].(string); ok {
				authConfig.Username = username
			}
			if password, ok := destRegistryPayload["password"].(string); ok {
				authConfig.Password = password
			}

			if authConfig.Username != "" || authConfig.Password != "" {
				encodedJSON, err := json.Marshal(authConfig)
				if err != nil {
					return fmt.Errorf("failed to encode destination registry auth: %w", err)
				}
				pushOptions.RegistryAuth = base64.URLEncoding.EncodeToString(encodedJSON)
			}

			e.log.Debug("Pushing image to internal registry", slog.String("destination", destinationRef))

			pushReader, err := e.dockerClient.ImagePush(ctx, destinationRef, pushOptions)
			if err != nil {
				return fmt.Errorf("failed to push snapshot to internal registry: %w", err)
			}
			defer pushReader.Close()

			// Consume the output to wait for push to complete
			for {
				_, err := pushReader.Read(buf)
				if err != nil {
					if err == io.EOF {
						break
					}
					break
				}
			}
		}
	}

	// Increment snapshot count
	e.collector.IncrementSnapshots()

	e.log.Info("Snapshot pulled successfully",
		slog.String("source_image", sourceImage),
		slog.String("destination_ref", destinationRef))
	return nil
}

func (e *Executor) removeSnapshot(ctx context.Context, job *apiclient.Job) error {
	// Get snapshot ref from resourceId
	snapshotRef := job.GetResourceId()
	if snapshotRef == "" {
		// Fallback to payload for backwards compatibility
		payload := job.GetPayload()
		var ok bool
		snapshotRef, ok = payload["snapshotRef"].(string)
		if !ok {
			return fmt.Errorf("snapshotRef not found in job")
		}
	}

	e.log.Info("Removing snapshot", slog.String("snapshot_ref", snapshotRef))

	// Remove the image from local storage
	_, err := e.dockerClient.ImageRemove(ctx, snapshotRef, image.RemoveOptions{
		Force:         true,
		PruneChildren: true,
	})
	if err != nil {
		// Log but don't fail if image doesn't exist
		e.log.Warn("Failed to remove snapshot image (may not exist)",
			slog.String("snapshot_ref", snapshotRef),
			slog.String("error", err.Error()))
	}

	// Decrement snapshot count
	e.collector.DecrementSnapshots()

	e.log.Info("Snapshot removed", slog.String("snapshot_ref", snapshotRef))
	return nil
}
