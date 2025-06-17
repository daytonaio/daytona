// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"context"
	"strings"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/pkg/common"
	"github.com/docker/docker/api/types/image"
)

func (s *SnapshotService) SnapshotExists(ctx context.Context, req *pb.SnapshotExistsRequest) (*pb.SnapshotExistsResponse, error) {
	snapshotName := strings.Replace(req.Snapshot, "docker.io/", "", 1)

	if strings.HasSuffix(snapshotName, ":latest") && !req.IncludeLatest {
		return &pb.SnapshotExistsResponse{
			Exists: false,
		}, nil
	}

	snapshots, err := s.dockerClient.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return &pb.SnapshotExistsResponse{
			Exists: false,
		}, common.MapDockerError(err)
	}

	found := false
	for _, snapshot := range snapshots {
		for _, tag := range snapshot.RepoTags {
			if strings.HasPrefix(tag, snapshotName) {
				found = true
				break
			}
		}
	}

	if found {
		s.log.Debug("Image already pulled", "snapshotName", snapshotName)
	}

	return &pb.SnapshotExistsResponse{
		Exists: found,
	}, nil
}
