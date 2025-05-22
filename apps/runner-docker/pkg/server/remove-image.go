// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"context"
	"fmt"

	pb "github.com/daytonaio/runner/proto"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/errdefs"
	log "github.com/sirupsen/logrus"
)

func (s *RunnerServer) RemoveImage(ctx context.Context, req *pb.RemoveImageRequest) (*pb.RemoveImageResponse, error) {
	_, err := s.dockerClient.ImageRemove(ctx, req.Image, image.RemoveOptions{
		Force:         req.Force,
		PruneChildren: true,
	})
	if err != nil {
		if errdefs.IsNotFound(err) {
			log.Infof("Image %s already removed and not found", req.Image)
			return &pb.RemoveImageResponse{
				Message: fmt.Sprintf("Image %s already removed or not found", req.Image),
			}, nil
		}
		return nil, err
	}

	log.Infof("Image %s removed successfully", req.Image)

	return &pb.RemoveImageResponse{
		Message: fmt.Sprintf("Image %s removed", req.Image),
	}, nil
}
