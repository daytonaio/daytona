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
	"github.com/daytonaio/runner-docker/pkg/common"
	pb "github.com/daytonaio/runner/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

func (s *RunnerServer) GetBuildLogs(req *pb.GetBuildLogsRequest, stream pb.Runner_GetBuildLogsServer) error {
	logFilePath, err := config.GetBuildLogFilePath(req.GetImageRef())
	if err != nil {
		return err
	}

	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		return common.NewNotFoundError(fmt.Errorf("build logs not found for ref: %s", req.ImageRef))
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

		return stream.Send(&pb.LogLine{Content: "Fetching build logs finished"})
	}

	reader := bufio.NewReader(file)

	checkImageRef := req.ImageRef

	// Fixed tag for instances where we are not looking for an entry with image ID
	if strings.HasPrefix(req.ImageRef, "daytona") {
		checkImageRef = req.ImageRef + ":daytona"
	}

	// flusher, ok := ctx.Writer.(http.Flusher)
	// if !ok {
	// 	return common.NewCustomError(http.StatusInternalServerError, "Streaming not supported", "STREAMING_NOT_SUPPORTED")
	// }

	go func() {
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				log.Errorf("Error reading log file: %v", err)
				break
			}

			if len(line) > 0 {
				err := stream.Send(&pb.LogLine{Content: string(line)})
				if err != nil {
					log.Errorf("Error writing to response: %v", err)
					break
				}
			}
		}
	}()

	for {
		existsResp, err := s.ImageExists(stream.Context(), &pb.ImageExistsRequest{
			Image:         checkImageRef,
			IncludeLatest: false,
		})
		if err != nil {
			log.Errorf("Error checking build status: %v", err)
			break
		}

		if existsResp.Exists {
			// If image exists, build is complete, allow time for the last logs to be written and break the loop
			time.Sleep(1 * time.Second)
			break
		}

		time.Sleep(250 * time.Millisecond)
	}

	return stream.Send(&pb.LogLine{
		Content: "Build completed successfully",
	})
}
