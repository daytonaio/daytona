// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"context"
	"fmt"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/pkg/services/common"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/errdefs"
	log "github.com/sirupsen/logrus"
)

func (s *SnapshotService) RemoveSnapshot(ctx context.Context, req *pb.RemoveSnapshotRequest) (*pb.RemoveSnapshotResponse, error) {
	_, err := s.dockerClient.ImageRemove(ctx, req.Snapshot, image.RemoveOptions{
		Force:         req.Force,
		PruneChildren: true,
	})
	if err != nil {
		if errdefs.IsNotFound(err) {
			log.Infof("Snapshot %s already removed and not found", req.GetSnapshot())
			return &pb.RemoveSnapshotResponse{
				Message: fmt.Sprintf("Snapshot %s already removed or not found", req.GetSnapshot()),
			}, nil
		}
		return nil, common.MapDockerError(err)
	}

	log.Infof("Snapshot %s removed successfully", req.GetSnapshot())

	return &pb.RemoveSnapshotResponse{
		Message: fmt.Sprintf("Snapshot %s removed", req.GetSnapshot()),
	}, nil
}
