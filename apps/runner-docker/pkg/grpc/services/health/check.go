// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package health

import (
	"context"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
)

func (s *HealthService) HealthCheck(ctx context.Context, in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status:  pb.HealthStatus_HEALTH_STATUS_HEALTHY,
		Version: "0.0.0-dev",
	}, nil
}
