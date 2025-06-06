// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/daytonaio/runner-docker/cmd/config"
	pb "github.com/daytonaio/runner/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BuildLogs implements the gRPC streaming method for build logs
func (s *RunnerServer) BuildLogs(req *pb.BuildLogsRequest, stream pb.Runner_BuildLogsServer) error {
	imageRef := req.GetImageRef()
	if imageRef == "" {
		return status.Error(codes.InvalidArgument, "imageRef parameter is required")
	}

	follow := req.GetFollow()

	logFilePath, err := config.GetBuildLogFilePath(imageRef)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		return status.Error(codes.NotFound, fmt.Sprintf("build logs not found for ref: %s", imageRef))
	}

	file, err := os.Open(logFilePath)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	defer file.Close()

	// If not following, just return the entire file content
	if !follow {
		data, err := io.ReadAll(file)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		return stream.Send(&pb.BuildLogsResponse{
			Data: data,
		})
	}

	// For streaming logs
	reader := bufio.NewReader(file)

	checkImageRef := imageRef

	// Fixed tag for instances where we are not looking for an entry with image ID
	if strings.HasPrefix(imageRef, "daytona") {
		checkImageRef = imageRef + ":daytona"
	}

	// Channel to coordinate between goroutines
	done := make(chan bool)
	errorChan := make(chan error, 1)

	// Start streaming existing log content
	go func() {
		defer func() {
			done <- true
		}()

		for {
			select {
			case <-stream.Context().Done():
				return
			default:
				line, err := reader.ReadBytes('\n')
				if err != nil && err != io.EOF {
					errorChan <- err
					return
				}

				if len(line) > 0 {
					if err := stream.Send(&pb.BuildLogsResponse{
						Data: line,
					}); err != nil {
						errorChan <- err
						return
					}
				}

				if err == io.EOF {
					done <- true
					return
				}
			}
		}
	}()

	// Monitor build completion
	go func() {
		for {
			select {
			case <-stream.Context().Done():
				return
			case <-done:
				return
			default:
				existsResponse, err := s.ImageExists(stream.Context(), &pb.ImageExistsRequest{
					Image:         checkImageRef,
					IncludeLatest: false,
				})
				if err != nil {
					log.Errorf("Error checking build status: %v", err)
					return
				}

				if existsResponse.Exists {
					// If image exists, build is complete, allow time for the last logs to be written
					time.Sleep(1 * time.Second)
					done <- true
					return
				}

				time.Sleep(250 * time.Millisecond)
			}
		}
	}()

	// Wait for completion or error
	select {
	case <-done:
		return nil
	case err := <-errorChan:
		return status.Error(codes.Internal, err.Error())
	case <-stream.Context().Done():
		return stream.Context().Err()
	}
}
