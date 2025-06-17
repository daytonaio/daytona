// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"context"
	"strings"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/pkg/services/common"
	"github.com/docker/docker/api/types/image"
	log "github.com/sirupsen/logrus"
)

func (s *SnapshotService) SnapshotExists(ctx context.Context, req *pb.SnapshotExistsRequest) (*pb.SnapshotExistsResponse, error) {
	imageName := strings.Replace(req.Snapshot, "docker.io/", "", 1)

	if strings.HasSuffix(imageName, ":latest") && !req.IncludeLatest {
		return &pb.SnapshotExistsResponse{
			Exists: false,
		}, nil
	}

	images, err := s.dockerClient.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return &pb.SnapshotExistsResponse{
			Exists: false,
		}, common.MapDockerError(err)
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
		log.Debugf("Image %s already pulled", imageName)
	}

	return &pb.SnapshotExistsResponse{
		Exists: found,
	}, nil
}
