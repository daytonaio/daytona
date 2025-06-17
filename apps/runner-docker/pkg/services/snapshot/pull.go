// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"context"
	"fmt"
	"io"
	"strings"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/internal/constants"
	"github.com/daytonaio/runner-docker/internal/util"
	"github.com/daytonaio/runner-docker/pkg/services/common"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/jsonmessage"
	log "github.com/sirupsen/logrus"
)

func (s *SnapshotService) PullSnapshot(ctx context.Context, req *pb.PullSnapshotRequest) (*pb.PullSnapshotResponse, error) {
	tag := "latest"
	lastColonIndex := strings.LastIndex(req.GetSnapshot(), ":")
	if lastColonIndex != -1 {
		tag = req.GetSnapshot()[lastColonIndex+1:]
	}

	if tag != "latest" {
		resp, err := s.SnapshotExists(ctx, &pb.SnapshotExistsRequest{
			Snapshot:      req.GetSnapshot(),
			IncludeLatest: true,
		})
		if err != nil {
			return nil, err
		}

		if resp.Exists {
			return &pb.PullSnapshotResponse{
				Message: fmt.Sprintf("Snapshot %s already pulled", req.GetSnapshot()),
			}, nil
		}
	}

	log.Infof("Pulling snapshot %s...", req.GetSnapshot())

	sandboxIdValue := ctx.Value(constants.ID_KEY)

	if sandboxIdValue != nil {
		sandboxId := sandboxIdValue.(string)
		s.cache.SetSandboxState(ctx, sandboxId, pb.SandboxState_SANDBOX_STATE_PULLING_SNAPSHOT)
	}

	responseBody, err := s.dockerClient.ImagePull(ctx, req.GetSnapshot(), image.PullOptions{
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

	log.Infof("Snapshot %s pulled successfully", req.GetSnapshot())

	return &pb.PullSnapshotResponse{
		Message: fmt.Sprintf("Snapshot %s pulled", req.GetSnapshot()),
	}, nil
}
