// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package health

import (
	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
)

type HealthService struct {
	pb.UnimplementedHealthServiceServer
}

// new service
func NewHealthService() *HealthService {
	return &HealthService{}
}
