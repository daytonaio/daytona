// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/daytonaio/common-go/pkg/log"
	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/storage"

	"github.com/docker/docker/api/types/build"
	docker_registry "github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/pkg/jsonmessage"
)

func (d *DockerClient) BuildImage(ctx context.Context, buildImageDto dto.BuildSnapshotRequestDTO) error {
	if !strings.Contains(buildImageDto.Snapshot, ":") || strings.HasSuffix(buildImageDto.Snapshot, ":") {
		return fmt.Errorf("invalid image format: must contain exactly one colon (e.g., 'myimage:1.0')")
	}

	d.logger.InfoContext(ctx, "Building image")

	// Check if image already exists
	exists, err := d.ImageExists(ctx, buildImageDto.Snapshot, true)
	if err != nil {
		return fmt.Errorf("failed to check if image exists: %w", err)
	}
	if exists {
		d.logger.InfoContext(ctx, "Image already built")
		return nil
	}

	// Create a build context from the provided hashes
	buildContextTar := new(bytes.Buffer)
	tarWriter := tar.NewWriter(buildContextTar)
	defer tarWriter.Close()

	dockerfileContent := []byte(buildImageDto.Dockerfile)
	dockerfileHeader := &tar.Header{
		Name: "Dockerfile",
		Mode: 0644,
		Size: int64(len(dockerfileContent)),
	}
	if err := tarWriter.WriteHeader(dockerfileHeader); err != nil {
		return fmt.Errorf("failed to write Dockerfile header to tar: %w", err)
	}
	if _, err := tarWriter.Write(dockerfileContent); err != nil {
		return fmt.Errorf("failed to write Dockerfile to tar: %w", err)
	}

	// Add context files if provided
	if len(buildImageDto.Context) > 0 {
		storageClient, err := storage.GetObjectStorageClient()
		if err != nil {
			return fmt.Errorf("failed to initialize object storage client: %w", err)
		}

		// Process each hash and extract the corresponding tar file
		for _, hash := range buildImageDto.Context {
			// Get the tar file from object storage
			tarData, err := storageClient.GetObject(ctx, buildImageDto.OrganizationId, hash)
			if err != nil {
				return fmt.Errorf("failed to get tar from storage with hash %s: %w", hash, err)
			}

			d.logger.InfoContext(ctx, "Processing context file with hash", "hash", hash, "bytes", len(tarData))

			if len(tarData) == 0 {
				return fmt.Errorf("empty tar file received for hash %s", hash)
			}

			tarReader := tar.NewReader(bytes.NewReader(tarData))

			// Extract each file from the tar and add it to the build context
			for {
				header, err := tarReader.Next()
				if err == io.EOF {
					break // End of tar archive
				}
				if err != nil {
					d.logger.WarnContext(ctx, "Error reading tar with hash", "hash", hash, "error", err)
					// Skip this tar file and continue with the next one
					break
				}

				// Skip directories
				if header.Typeflag == tar.TypeDir {
					continue
				}

				fileContent := new(bytes.Buffer)
				_, err = io.Copy(fileContent, tarReader)
				if err != nil {
					d.logger.WarnContext(ctx, "Failed to read file from tar", "file", header.Name, "error", err)
					continue // Skip this file and continue with the next one
				}

				buildHeader := &tar.Header{
					Name: header.Name,
					Mode: header.Mode,
					Size: int64(fileContent.Len()),
				}

				if err := tarWriter.WriteHeader(buildHeader); err != nil {
					d.logger.WarnContext(ctx, "Failed to write tar header", "file", header.Name, "error", err)
					continue
				}

				if _, err := tarWriter.Write(fileContent.Bytes()); err != nil {
					d.logger.WarnContext(ctx, "Failed to write file to tar", "file", header.Name, "error", err)
					continue
				}

				d.logger.InfoContext(ctx, "Added file to build context", "file", header.Name)
			}
		}
	}

	buildContext := io.NopCloser(buildContextTar)

	var authConfigs map[string]docker_registry.AuthConfig

	if len(buildImageDto.SourceRegistries) > 0 {
		authConfigs = make(map[string]docker_registry.AuthConfig, len(buildImageDto.SourceRegistries)*2)
		for _, sourceRegistry := range buildImageDto.SourceRegistries {
			if !sourceRegistry.HasAuth() {
				continue
			}

			authConfig := docker_registry.AuthConfig{
				Username: *sourceRegistry.Username,
				Password: *sourceRegistry.Password,
			}

			url := sourceRegistry.Url
			url = strings.TrimPrefix(url, "https://")
			url = strings.TrimPrefix(url, "http://")

			// Docker's build API expects 'index.docker.io/v1/' for Docker Hub auth
			if url == "docker.io" {
				authConfigs["https://index.docker.io/v1/"] = authConfig
			} else {
				authConfigs["https://"+url] = authConfig
				authConfigs["http://"+url] = authConfig
			}
		}
	}

	resp, err := d.apiClient.ImageBuild(ctx, buildContext, build.ImageBuildOptions{
		Tags:        []string{buildImageDto.Snapshot},
		Dockerfile:  "Dockerfile",
		Remove:      true,
		ForceRemove: true,
		PullParent:  true,
		Platform:    "linux/amd64", // Force AMD64 architecture
		AuthConfigs: authConfigs,
	})
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	logFilePath, err := config.GetBuildLogFilePath(buildImageDto.Snapshot)
	if err != nil {
		return err
	}

	var writer io.Writer = &log.DebugLogWriter{}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		d.logger.ErrorContext(ctx, "Failed to open log file", "error", err)
	} else {
		defer logFile.Close()
		writer = io.MultiWriter(&log.DebugLogWriter{}, logFile)
	}

	err = jsonmessage.DisplayJSONMessagesStream(resp.Body, writer, 0, true, nil)
	if err != nil {
		return err
	}

	d.logger.InfoContext(ctx, "Image built successfully")

	return nil
}
