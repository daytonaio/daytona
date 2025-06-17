// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/pkg/common"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type VersionResponse struct {
	Version string `json:"version"`
}

func (s *SandboxService) SandboxDaemonVersion(ctx context.Context, req *pb.SandboxDaemonVersionRequest) (*pb.SandboxDaemonVersionResponse, error) {
	// Get container details using Docker API client
	c, err := s.dockerClient.ContainerInspect(ctx, req.GetSandboxId())
	if err != nil {
		return nil, common.MapDockerError(err)
	}

	// Extract container IP from network settings
	var containerIP string
	for _, network := range c.NetworkSettings.Networks {
		containerIP = network.IPAddress
		break
	}

	if containerIP == "" {
		return nil, status.Errorf(codes.InvalidArgument, "container has no IP address, it might not be running")
	}

	// Make HTTP GET request to toolbox version endpoint
	versionURL := fmt.Sprintf("http://%s:2280/version", containerIP)
	resp, err := http.Get(versionURL)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get daemon version: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, status.Errorf(codes.Internal, "failed to get daemon version, status: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read response body: %v", err)
	}

	// Parse the JSON response to extract version
	var versionResp VersionResponse
	if err := json.Unmarshal(body, &versionResp); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse version response: %v", err)
	}

	return &pb.SandboxDaemonVersionResponse{
		DaemonVersion: versionResp.Version,
	}, nil
}
