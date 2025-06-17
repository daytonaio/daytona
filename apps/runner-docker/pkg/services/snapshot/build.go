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

	"github.com/daytonaio/runner-docker/cmd/config"
	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/internal/util"
	"github.com/daytonaio/runner-docker/pkg/services/common"
	"github.com/daytonaio/runner-docker/pkg/storage"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/jsonmessage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *SnapshotService) BuildSnapshot(ctx context.Context, req *pb.BuildSnapshotRequest) (*pb.BuildSnapshotResponse, error) {
	if !strings.Contains(req.GetSnapshot(), ":") || strings.HasSuffix(req.GetSnapshot(), ":") {
		return nil, status.Errorf(codes.InvalidArgument, "snapshot name must include a valid tag")
	}

	if !strings.Contains(req.GetSnapshot(), ":") || strings.HasSuffix(req.GetSnapshot(), ":") {
		return nil, status.Errorf(codes.InvalidArgument, "snapshot name must include a valid tag")
	}

	if s.logWriter != nil {
		s.logWriter.Write([]byte("Building image...\n"))
	}

	// Check if image already exists
	existsResp, err := s.SnapshotExists(ctx, &pb.SnapshotExistsRequest{
		Snapshot:      req.GetSnapshot(),
		IncludeLatest: true,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check if image exists: %w", err)
	}
	if existsResp.Exists {
		if s.logWriter != nil {
			s.logWriter.Write([]byte("Image already built\n"))
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
		return nil, status.Errorf(codes.Internal, "failed to write Dockerfile header to tar: %w", err)
	}
	if _, err := tarWriter.Write(dockerfileContent); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to write Dockerfile to tar: %w", err)
	}

	// Add context files if provided
	if len(req.Context) > 0 {
		storageClient, err := storage.GetObjectStorageClient()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to initialize object storage client: %w", err)
		}

		// Process each hash and extract the corresponding tar file
		for _, hash := range req.Context {
			// Get the tar file from object storage
			tarData, err := storageClient.GetObject(ctx, req.OrganizationId, hash)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to get tar from storage with hash %s: %w", hash, err)
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

	// Extract image name without tag
	filePath := req.GetSnapshot()[:strings.LastIndex(req.GetSnapshot(), ":")]

	logFilePath, err := config.GetBuildLogFilePath(filePath)
	if err != nil {
		return nil, err
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
	}
	defer logFile.Close()

	multiWriter := io.MultiWriter(s.logWriter, logFile)

	err = jsonmessage.DisplayJSONMessagesStream(resp.Body, multiWriter, 0, true, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("failed to stream build output: %w", err).Error())
	}

	tag := req.GetSnapshot()

	if req.GetPushToInternalRegistry() {
		registry := req.GetRegistry()
		if registry.GetProject() == "" {
			return nil, status.Errorf(codes.InvalidArgument, "project is required when pushing to internal registry")
		}
		tag = fmt.Sprintf("%s/%s/%s", registry.GetUrl(), registry.GetProject(), req.GetSnapshot())
	}

	err = s.tagImage(ctx, req.GetSnapshot(), tag)
	if err != nil {
		return nil, err
	}

	if req.GetPushToInternalRegistry() {
		if s.logWriter != nil {
			s.logWriter.Write([]byte(fmt.Sprintf("Pushing image %s...", tag)))
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
			s.logWriter.Write([]byte(fmt.Sprintf("Image %s pushed successfully", tag)))
		}
	}

	if s.logWriter != nil {
		s.logWriter.Write([]byte("Image built successfully\n"))
	}

	return &pb.BuildSnapshotResponse{
		Message: "Image built successfully",
	}, nil
}

func (s *SnapshotService) tagImage(ctx context.Context, sourceImage string, targetImage string) error {
	if s.logWriter != nil {
		s.logWriter.Write([]byte(fmt.Sprintf("Tagging image %s as %s...\n", sourceImage, targetImage)))
	}

	// Extract repository and tag from targetImage
	lastColonIndex := strings.LastIndex(targetImage, ":")
	var repo, tag string

	if lastColonIndex == -1 {
		return status.Errorf(codes.InvalidArgument, "invalid target image format: %s", targetImage)
	} else {
		repo = targetImage[:lastColonIndex]
		tag = targetImage[lastColonIndex+1:]
	}

	if repo == "" || tag == "" {
		return status.Errorf(codes.InvalidArgument, "invalid target image format: %s", targetImage)
	}

	err := s.dockerClient.ImageTag(ctx, sourceImage, targetImage)
	if err != nil {
		return common.MapDockerError(err)
	}

	if s.logWriter != nil {
		s.logWriter.Write([]byte(fmt.Sprintf("Image tagged successfully: %s → %s\n", sourceImage, targetImage)))
	}

	return nil
}
