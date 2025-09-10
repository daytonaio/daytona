// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/internal/util"
	"github.com/daytonaio/runner-docker/pkg/common"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/jsonmessage"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type backupContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

var backup_context_map = cmap.New[backupContext]()

func (s *SandboxService) CreateBackup(ctx context.Context, req *pb.CreateBackupRequest) (*pb.CreateBackupResponse, error) {
	// Cancel a backup if it's already in progress
	backup_context, ok := backup_context_map.Get(req.GetSandboxId())
	if ok {
		backup_context.cancel()
	}

	s.log.Info("Creating backup for container %s...", "sandboxId", req.GetSandboxId())

	s.cache.SetBackupState(ctx, req.GetSandboxId(), pb.BackupState_BACKUP_STATE_IN_PROGRESS, nil)

	go func() {
		ctx, cancel := context.WithCancel(context.Background())

		defer func() {
			backupContext, ok := backup_context_map.Get(req.GetSandboxId())
			if ok {
				backupContext.cancel()
			}
			backup_context_map.Remove(req.GetSandboxId())
		}()

		backup_context_map.Set(req.GetSandboxId(), backupContext{ctx, cancel})

		err := s.commitContainer(ctx, req.GetSandboxId(), req.GetSnapshot())
		if err != nil {
			if errors.Is(err, context.Canceled) {
				s.cache.SetBackupState(ctx, req.GetSandboxId(), pb.BackupState_BACKUP_STATE_UNSPECIFIED, nil)
				s.log.Info("Backup for container %s canceled", "sandboxId", req.GetSandboxId())
				return
			}
			s.log.Error("Error committing container %s: %v", "sandboxId", req.GetSandboxId(), "error", err)
			s.cache.SetBackupState(ctx, req.GetSandboxId(), pb.BackupState_BACKUP_STATE_FAILED, err)
			return
		}

		err = s.pushImage(ctx, req.GetSnapshot(), req.GetRegistry())
		if err != nil {
			if errors.Is(err, context.Canceled) {
				s.cache.SetBackupState(ctx, req.GetSandboxId(), pb.BackupState_BACKUP_STATE_UNSPECIFIED, nil)
				s.log.Info("Backup for container %s canceled", "sandboxId", req.GetSandboxId())
				return
			}
			s.log.Error("Error pushing image %s: %v", "snapshot", req.GetSnapshot(), "error", err)
			s.cache.SetBackupState(ctx, req.GetSandboxId(), pb.BackupState_BACKUP_STATE_FAILED, err)
			return
		}

		s.cache.SetBackupState(ctx, req.GetSandboxId(), pb.BackupState_BACKUP_STATE_COMPLETED, nil)

		s.log.Info("Backup (%s) for container %s created successfully", "snapshot", req.GetSnapshot(), "sandboxId", req.GetSandboxId())

		_, err = s.dockerClient.ImageRemove(ctx, req.GetSnapshot(), image.RemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
		if err != nil {
			if errdefs.IsNotFound(err) {
				s.log.Info("Image %s already removed and not found", "snapshot", req.GetSnapshot())
			} else {
				s.log.Error("Error removing image %s: %v", "snapshot", req.GetSnapshot(), "error", err)
				// Don't set backup to failed because the image is already pushed
			}
		} else {
			s.log.Info("Image %s deleted successfully", "snapshot", req.GetSnapshot())
		}
	}()

	return &pb.CreateBackupResponse{
		Message: fmt.Sprintf("Backup created for sandbox %s", req.GetSandboxId()),
	}, nil
}

func (s *SandboxService) commitContainer(ctx context.Context, sandboxId, snapshot string) error {
	const maxRetries = 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		s.log.Info("Committing container %s (attempt %d/%d)...", "sandboxId", sandboxId, "attempt", attempt, "maxRetries", maxRetries)

		commitResp, err := s.dockerClient.ContainerCommit(ctx, sandboxId, container.CommitOptions{
			Reference: snapshot,
			Pause:     false,
		})
		if err == nil {
			s.log.Info("Container %s committed successfully with image ID: %s", "sandboxId", sandboxId, "imageId", commitResp.ID)
			return nil
		}

		// Check if the error is related to "failed to get digest" and try export/import fallback
		if strings.Contains(err.Error(), "Error response from daemon: failed to get digest") {
			s.log.Warn("Commit failed with digest error, attempting export/import fallback for container %s", "sandboxId", sandboxId)

			err = s.exportImportContainer(ctx, sandboxId, snapshot)
			if err == nil {
				s.log.Info("Container %s successfully backed up using export/import method", "sandboxId", sandboxId)
				return nil
			}

			s.log.Error("Export/import fallback also failed for container %s: %v", "sandboxId", sandboxId, "error", err)
		}

		if attempt < maxRetries {
			s.log.Warn("Failed to commit container %s (attempt %d/%d): %v", "sandboxId", sandboxId, "attempt", attempt, "maxRetries", maxRetries, "error", err)
			continue
		}

		return fmt.Errorf("failed to commit container after %d attempts: %w", maxRetries, err)
	}

	return nil
}

func (s *SandboxService) exportImportContainer(ctx context.Context, sandboxId, imageName string) error {
	s.log.Info("Exporting container %s and importing as image %s...", "sandboxId", sandboxId, "imageName", imageName)

	// First, inspect the container to get its configuration
	containerInfo, err := s.dockerClient.ContainerInspect(ctx, sandboxId)
	if err != nil {
		return fmt.Errorf("failed to inspect container %s: %w", sandboxId, err)
	}

	// Export the container
	exportReader, err := s.dockerClient.ContainerExport(ctx, sandboxId)
	if err != nil {
		return fmt.Errorf("failed to export container %s: %w", sandboxId, err)
	}
	defer exportReader.Close()

	// Prepare import options with container configuration
	importOptions := image.ImportOptions{
		Message: fmt.Sprintf("Imported from container %s", sandboxId),
	}

	// Build the configuration changes to preserve CMD, ENTRYPOINT, ENV, etc.
	var changes []string

	// Preserve CMD if it exists
	if len(containerInfo.Config.Cmd) > 0 {
		cmdStr := buildDockerfileCmd(containerInfo.Config.Cmd)
		changes = append(changes, fmt.Sprintf("CMD %s", cmdStr))
	}

	// Preserve ENTRYPOINT if it exists
	if len(containerInfo.Config.Entrypoint) > 0 {
		entrypointStr := buildDockerfileCmd(containerInfo.Config.Entrypoint)
		changes = append(changes, fmt.Sprintf("ENTRYPOINT %s", entrypointStr))
	}

	// Preserve environment variables
	if len(containerInfo.Config.Env) > 0 {
		for _, env := range containerInfo.Config.Env {
			changes = append(changes, fmt.Sprintf("ENV %s", env))
		}
	}

	// Preserve working directory
	if containerInfo.Config.WorkingDir != "" {
		changes = append(changes, fmt.Sprintf("WORKDIR %s", containerInfo.Config.WorkingDir))
	}

	// Preserve exposed ports
	if len(containerInfo.Config.ExposedPorts) > 0 {
		for port := range containerInfo.Config.ExposedPorts {
			changes = append(changes, fmt.Sprintf("EXPOSE %s", string(port)))
		}
	}

	// Preserve user
	if containerInfo.Config.User != "" {
		changes = append(changes, fmt.Sprintf("USER %s", containerInfo.Config.User))
	}

	// Apply the changes
	importOptions.Changes = changes

	s.log.Info("Applying configuration changes: %v", "changes", changes)

	importResponse, err := s.dockerClient.ImageImport(ctx, image.ImportSource{
		Source:     exportReader,
		SourceName: "-",
	}, imageName, importOptions)
	if err != nil {
		return fmt.Errorf("failed to import container %s as image %s: %w", sandboxId, imageName, err)
	}
	defer importResponse.Close()

	// Read the import response to completion
	_, err = io.ReadAll(importResponse)
	if err != nil {
		return fmt.Errorf("failed to read import response for container %s: %w", sandboxId, err)
	}

	s.log.Info("Container %s successfully exported and imported as image %s with preserved configuration", "sandboxId", sandboxId, "imageName", imageName)
	return nil
}

// buildDockerfileCmd converts a slice of command arguments to a properly formatted Dockerfile CMD/ENTRYPOINT string
func buildDockerfileCmd(cmd []string) string {
	if len(cmd) == 0 {
		return ""
	}

	// Use JSON array format for better compatibility
	var quotedArgs []string
	for _, arg := range cmd {
		// Escape quotes and backslashes in the argument
		escaped := strings.ReplaceAll(arg, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		quotedArgs = append(quotedArgs, fmt.Sprintf("\"%s\"", escaped))
	}

	return fmt.Sprintf("[%s]", strings.Join(quotedArgs, ", "))
}

func (s *SandboxService) pushImage(ctx context.Context, snapshot string, registry *pb.Registry) error {
	s.log.Info("Pushing image %s...", "snapshot", snapshot)

	responseBody, err := s.dockerClient.ImagePush(ctx, snapshot, image.PushOptions{
		RegistryAuth: common.GetRegistryAuth(registry),
	})
	if err != nil {
		return common.MapDockerError(err)
	}
	defer responseBody.Close()

	err = jsonmessage.DisplayJSONMessagesStream(responseBody, io.Writer(&util.DebugLogWriter{}), 0, true, nil)
	if err != nil {
		return err
	}

	s.log.Info("Image %s pushed successfully", "snapshot", snapshot)

	return nil
}
