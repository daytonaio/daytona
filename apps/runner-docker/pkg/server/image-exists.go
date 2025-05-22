// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"context"
	"strings"

	pb "github.com/daytonaio/runner/proto"
	"github.com/docker/docker/api/types/image"
	log "github.com/sirupsen/logrus"
)

func (s *RunnerServer) ImageExists(ctx context.Context, req *pb.ImageExistsRequest) (*pb.ImageExistsResponse, error) {
	imageName := strings.Replace(req.Image, "docker.io/", "", 1)

	if strings.HasSuffix(imageName, ":latest") && !req.IncludeLatest {
		return &pb.ImageExistsResponse{
			Exists: false,
		}, nil
	}

	images, err := s.dockerClient.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return &pb.ImageExistsResponse{
			Exists: false,
		}, err
	}

	found := false
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if strings.HasPrefix(tag, imageName) {
				found = true
				break
			}
		}
	}

	if found {
		log.Infof("Image %s already pulled", imageName)
	}

	return &pb.ImageExistsResponse{
		Exists: found,
	}, nil
}
