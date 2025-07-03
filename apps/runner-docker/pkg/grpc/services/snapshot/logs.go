// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"bufio"
	"io"
	"os"
	"strings"
	"time"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/internal/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *SnapshotService) GetSnapshotLogs(req *pb.GetSnapshotLogsRequest, stream pb.SnapshotService_GetSnapshotLogsServer) error {
	logFilePath, err := config.GetBuildLogFilePath(req.GetSnapshotRef())
	if err != nil {
		return err
	}

	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		return status.Errorf(codes.NotFound, "build logs not found for ref: %s", req.GetSnapshotRef())
	}

	err = stream.SetHeader(metadata.MD{
		"Content-Type": []string{"application/octet-stream"},
	})
	if err != nil {
		return err
	}

	file, err := os.Open(logFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// If not following, just return the entire file content
	if !req.GetFollow() {
		err = stream.SendMsg(file)
		if err != nil {
			return err
		}

		return stream.Send(&pb.GetSnapshotLogsResponse{Content: "Fetching build logs finished"})
	}

	reader := bufio.NewReader(file)

	checkSnapshotRef := req.GetSnapshotRef()

	// Fixed tag for instances where we are not looking for an entry with image ID
	if strings.HasPrefix(req.GetSnapshotRef(), "daytona") {
		checkSnapshotRef = req.GetSnapshotRef() + ":daytona"
	}

	// flusher, ok := ctx.Writer.(http.Flusher)
	// if !ok {
	// 	return common.NewCustomError(http.StatusInternalServerError, "Streaming not supported", "STREAMING_NOT_SUPPORTED")
	// }

	go func() {
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				s.log.Error("Error reading log file", "error", err)
				break
			}

			if len(line) > 0 {
				err := stream.Send(&pb.GetSnapshotLogsResponse{Content: string(line)})
				if err != nil {
					s.log.Error("Error writing to response", "error", err)
					break
				}
			}
		}
	}()

	for {
		existsResp, err := s.SnapshotExists(stream.Context(), &pb.SnapshotExistsRequest{
			Snapshot: checkSnapshotRef,
		})
		if err != nil {
			s.log.Error("Error checking build status", "error", err)
			break
		}

		if existsResp.Exists {
			// If image exists, build is complete, allow time for the last logs to be written and break the loop
			time.Sleep(1 * time.Second)
			break
		}

		time.Sleep(250 * time.Millisecond)
	}

	return stream.Send(&pb.GetSnapshotLogsResponse{
		Content: "Build completed successfully",
	})
}
