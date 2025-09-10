// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
)

func (s *SandboxService) GetNetworkSettings(ctx context.Context, req *pb.GetNetworkSettingsRequest) (*pb.GetNetworkSettingsResponse, error) {
	return &pb.GetNetworkSettingsResponse{
		NetworkBlockAll:  nil,
		NetworkAllowList: nil,
	}, nil
}
