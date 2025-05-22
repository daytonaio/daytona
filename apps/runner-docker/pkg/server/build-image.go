// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/daytonaio/runner-docker/cmd/config"
	"github.com/daytonaio/runner-docker/pkg/common"
	"github.com/daytonaio/runner-docker/pkg/storage"
	pb "github.com/daytonaio/runner/proto"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
)

func (s *RunnerServer) BuildImage(ctx context.Context, req *pb.BuildImageRequest) (*pb.BuildImageResponse, error) {
	if !strings.Contains(req.GetImage(), ":") || strings.HasSuffix(req.GetImage(), ":") {
		return nil, common.NewBadRequestError(errors.New("invalid image format: must contain exactly one colon (e.g., 'myimage:1.0')"))
	}

	if s.logWriter != nil {
		s.logWriter.Write([]byte("Building image...\n"))
	}

	// Check if image already exists
	existsResp, err := s.ImageExists(ctx, &pb.ImageExistsRequest{
		Image:         req.GetImage(),
		IncludeLatest: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check if image exists: %w", err)
	}
	if existsResp.Exists {
		if s.logWriter != nil {
			s.logWriter.Write([]byte("Image already built\n"))
		}
		return &pb.BuildImageResponse{
			Message: "Image already built",
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
		return nil, fmt.Errorf("failed to write Dockerfile header to tar: %w", err)
	}
	if _, err := tarWriter.Write(dockerfileContent); err != nil {
		return nil, fmt.Errorf("failed to write Dockerfile to tar: %w", err)
	}

	// Add context files if provided
	if len(req.Context) > 0 {
		storageClient, err := storage.GetObjectStorageClient()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize object storage client: %w", err)
		}

		// Process each hash and extract the corresponding tar file
		for _, hash := range req.Context {
			// Get the tar file from object storage
			tarData, err := storageClient.GetObject(ctx, req.OrganizationId, hash)
			if err != nil {
				return nil, fmt.Errorf("failed to get tar from storage with hash %s: %w", hash, err)
			}

			if s.logWriter != nil {
				s.logWriter.Write(fmt.Appendf(nil, "Processing context file with hash %s (%d bytes)\n", hash, len(tarData)))
			}

			if len(tarData) == 0 {
				return nil, fmt.Errorf("empty tar file received for hash %s", hash)
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
		Tags:        []string{req.GetImage()},
		Dockerfile:  "Dockerfile",
		Remove:      true,
		ForceRemove: true,
		PullParent:  true,
		Platform:    "linux/amd64", // Force AMD64 architecture
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	// Extract image name without tag
	filePath := req.GetImage()[:strings.LastIndex(req.GetImage(), ":")]

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
		return nil, fmt.Errorf("failed to stream build output: %w", err)
	}

	tag := req.GetImage()

	if req.GetPushToInternalRegistry() {
		registry := req.GetRegistry()
		if registry.GetProject() == "" {
			return nil, common.NewBadRequestError(errors.New("project is required when pushing to internal registry"))
		}
		tag = fmt.Sprintf("%s/%s/%s", registry.GetUrl(), registry.GetProject(), req.GetImage())
	}

	err = s.tagImage(ctx, req.GetImage(), tag)
	if err != nil {
		return nil, err
	}

	if req.GetPushToInternalRegistry() {
		err = s.pushImage(ctx, tag, req.GetRegistry())
		if err != nil {
			return nil, err
		}
	}

	if s.logWriter != nil {
		s.logWriter.Write([]byte("Image built successfully\n"))
	}

	return &pb.BuildImageResponse{
		Message: "Image built successfully",
	}, nil
}

func (s *RunnerServer) tagImage(ctx context.Context, sourceImage string, targetImage string) error {
	if s.logWriter != nil {
		s.logWriter.Write([]byte(fmt.Sprintf("Tagging image %s as %s...\n", sourceImage, targetImage)))
	}

	// Extract repository and tag from targetImage
	lastColonIndex := strings.LastIndex(targetImage, ":")
	var repo, tag string

	if lastColonIndex == -1 {
		return fmt.Errorf("invalid target image format: %s", targetImage)
	} else {
		repo = targetImage[:lastColonIndex]
		tag = targetImage[lastColonIndex+1:]
	}

	if repo == "" || tag == "" {
		return fmt.Errorf("invalid target image format: %s", targetImage)
	}

	err := s.dockerClient.ImageTag(ctx, sourceImage, targetImage)
	if err != nil {
		return err
	}

	if s.logWriter != nil {
		s.logWriter.Write([]byte(fmt.Sprintf("Image tagged successfully: %s → %s\n", sourceImage, targetImage)))
	}

	return nil
}
