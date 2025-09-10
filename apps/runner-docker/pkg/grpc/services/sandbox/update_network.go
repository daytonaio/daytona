// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/pkg/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *SandboxService) UpdateNetworkSettings(ctx context.Context, req *pb.UpdateNetworkSettingsRequest) (*pb.UpdateNetworkSettingsResponse, error) {
	info, err := s.dockerClient.ContainerInspect(ctx, req.GetSandboxId())
	if err != nil {
		return nil, common.MapDockerError(err)
	}
	containerShortId := info.ID[:12]

	// Return error if container does not have an IP address
	if info.NetworkSettings.IPAddress == "" {
		return nil, status.Errorf(codes.InvalidArgument, "sandbox does not have an IP address")
	}

	if req.GetNetworkBlockAll() {
		err = s.netRulesManager.SetNetWorkRules(containerShortId, info.NetworkSettings.IPAddress, "")
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update network settings: %v", err)
		}
	} else if req.GetNetworkAllowList() != "" {
		err = s.netRulesManager.SetNetWorkRules(containerShortId, info.NetworkSettings.IPAddress, req.GetNetworkAllowList())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update network settings: %v", err)
		}
	}
	
	return &pb.UpdateNetworkSettingsResponse{
		Message: "Network settings updated",
	}, nil
}