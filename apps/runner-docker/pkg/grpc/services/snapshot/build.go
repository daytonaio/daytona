// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/internal/util"
	"github.com/daytonaio/runner-docker/pkg/common"
	"github.com/daytonaio/runner-docker/pkg/storage"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/jsonmessage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *SnapshotService) BuildSnapshot(ctx context.Context, req *pb.BuildSnapshotRequest) (*pb.BuildSnapshotResponse, error) {
	if !strings.Contains(req.GetSnapshot(), ":") || strings.HasSuffix(req.GetSnapshot(), ":") {
		return nil, status.Error(codes.InvalidArgument, "snapshot name must include a valid tag")
	}

	if s.logWriter != nil {
		s.logWriter.Write([]byte("Building snapshot...\n"))
	}

	// Check if snapshot already exists
	existsResp, err := s.SnapshotExists(ctx, &pb.SnapshotExistsRequest{
		Snapshot:      req.GetSnapshot(),
		IncludeLatest: true,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to check if snapshot exists: %w", err).Error())
	}
	if existsResp.Exists {
		if s.logWriter != nil {
			s.logWriter.Write([]byte("Snapshot already built\n"))
		}
		return &pb.BuildSnapshotResponse{
			Message: "Snapshot already built",
		}, nil
	}

	// Create a build context from the provided hashes
	buildContextTar := new(bytes.Buffer)
	tarWriter := tar.NewWriter(buildContextTar)
	defer tarWriter.Close()

	dockerfileContent := []byte(req.Dockerfile)
	dockerfileHeader := &tar.Header{
		Name: "Dockerfile",
		Mode: 0644,
		Size: int64(len(dockerfileContent)),
	}
	if err := tarWriter.WriteHeader(dockerfileHeader); err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to write Dockerfile header to tar: %w", err).Error())
	}
	if _, err := tarWriter.Write(dockerfileContent); err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to write Dockerfile to tar: %w", err).Error())
	}

	// Add context files if provided
	if len(req.Context) > 0 {
		storageClient, err := storage.GetObjectStorageClient()
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("failed to initialize object storage client: %w", err).Error())
		}

		// Process each hash and extract the corresponding tar file
		for _, hash := range req.Context {
			// Get the tar file from object storage
			tarData, err := storageClient.GetObject(ctx, req.OrganizationId, hash)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Errorf("failed to get tar from storage with hash %s: %w", hash, err).Error())
			}

			if s.logWriter != nil {
				s.logWriter.Write(fmt.Appendf(nil, "Processing context file with hash %s (%d bytes)\n", hash, len(tarData)))
			}

			if len(tarData) == 0 {
				return nil, status.Errorf(codes.Internal, "empty tar file received for hash %s", hash)
			}

			tarReader := tar.NewReader(bytes.NewReader(tarData))

			// Extract each file from the tar and add it to the build context
			for {
				header, err := tarReader.Next()
				if err == io.EOF {
					break // End of tar archive
				}
				if err != nil {
					if s.logWriter != nil {
						s.logWriter.Write([]byte(fmt.Sprintf("Warning: Error reading tar with hash %s: %v\n", hash, err)))
					}
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
					if s.logWriter != nil {
						s.logWriter.Write([]byte(fmt.Sprintf("Warning: Failed to read file %s from tar: %v\n", header.Name, err)))
					}
					continue // Skip this file and continue with the next one
				}

				buildHeader := &tar.Header{
					Name: header.Name,
					Mode: header.Mode,
					Size: int64(fileContent.Len()),
				}

				if err := tarWriter.WriteHeader(buildHeader); err != nil {
					if s.logWriter != nil {
						s.logWriter.Write([]byte(fmt.Sprintf("Warning: Failed to write tar header for %s: %v\n", header.Name, err)))
					}
					continue
				}

				if _, err := tarWriter.Write(fileContent.Bytes()); err != nil {
					if s.logWriter != nil {
						s.logWriter.Write([]byte(fmt.Sprintf("Warning: Failed to write file %s to tar: %v\n", header.Name, err)))
					}
					continue
				}

				if s.logWriter != nil {
					s.logWriter.Write([]byte(fmt.Sprintf("Added %s to build context\n", header.Name)))
				}
			}
		}
	}

	buildContext := io.NopCloser(buildContextTar)

	resp, err := s.dockerClient.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
		Tags:        []string{req.GetSnapshot()},
		Dockerfile:  "Dockerfile",
		Remove:      true,
		ForceRemove: true,
		PullParent:  true,
		Platform:    "linux/amd64", // Force AMD64 architecture
	})
	if err != nil {
		return nil, common.MapDockerError(err)
	}
	defer resp.Body.Close()

	// Extract snapshot name without tag
	filePath := req.GetSnapshot()[:strings.LastIndex(req.GetSnapshot(), ":")]

	logFilePath, err := s.getBuildLogFilePath(filePath)
	if err != nil {
		return nil, err
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		s.log.Error("Failed to open log file", "error", err)
	}
	defer logFile.Close()

	multiWriter := io.MultiWriter(s.logWriter, logFile)

	err = jsonmessage.DisplayJSONMessagesStream(resp.Body, multiWriter, 0, true, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to stream build output: %w", err).Error())
	}

	tag := req.GetSnapshot()

	if req.GetPushToInternalRegistry() {
		registry := req.GetRegistry()
		if registry.GetProject() == "" {
			return nil, status.Error(codes.InvalidArgument, "project is required when pushing to internal registry")
		}
		tag = fmt.Sprintf("%s/%s/%s", registry.GetUrl(), registry.GetProject(), req.GetSnapshot())
	}

	err = s.tagSnapshot(ctx, req.GetSnapshot(), tag)
	if err != nil {
		return nil, err
	}

	if req.GetPushToInternalRegistry() {
		if s.logWriter != nil {
			s.logWriter.Write([]byte(fmt.Sprintf("Pushing snapshot %s...", tag)))
		}

		responseBody, err := s.dockerClient.ImagePush(ctx, tag, image.PushOptions{
			RegistryAuth: common.GetRegistryAuth(req.GetRegistry()),
		})
		if err != nil {
			return nil, common.MapDockerError(err)
		}
		defer responseBody.Close()

		err = jsonmessage.DisplayJSONMessagesStream(responseBody, io.Writer(&util.DebugLogWriter{}), 0, true, nil)
		if err != nil {
			return nil, err
		}

		if s.logWriter != nil {
			s.logWriter.Write([]byte(fmt.Sprintf("Snapshot %s pushed successfully", tag)))
		}
	}

	if s.logWriter != nil {
		s.logWriter.Write([]byte("Snapshot built successfully\n"))
	}

	return &pb.BuildSnapshotResponse{
		Message: "Snapshot built successfully",
	}, nil
}

func (s *SnapshotService) tagSnapshot(ctx context.Context, sourceSnapshot string, targetSnapshot string) error {
	if s.logWriter != nil {
		s.logWriter.Write([]byte(fmt.Sprintf("Tagging snapshot %s as %s...\n", sourceSnapshot, targetSnapshot)))
	}

	// Extract repository and tag from targetSnapshot
	lastColonIndex := strings.LastIndex(targetSnapshot, ":")
	var repo, tag string

	if lastColonIndex == -1 {
		return status.Errorf(codes.InvalidArgument, "invalid target snapshot format: %s", targetSnapshot)
	} else {
		repo = targetSnapshot[:lastColonIndex]
		tag = targetSnapshot[lastColonIndex+1:]
	}

	if repo == "" || tag == "" {
		return status.Errorf(codes.InvalidArgument, "invalid target snapshot format: %s", targetSnapshot)
	}

	err := s.dockerClient.ImageTag(ctx, sourceSnapshot, targetSnapshot)
	if err != nil {
		return common.MapDockerError(err)
	}

	if s.logWriter != nil {
		s.logWriter.Write([]byte(fmt.Sprintf("Snapshot tagged successfully: %s → %s\n", sourceSnapshot, targetSnapshot)))
	}

	return nil
}
