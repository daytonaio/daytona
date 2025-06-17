// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package health

import (
	"log/slog"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
)

type HealthService struct {
	pb.UnimplementedHealthServiceServer
	log *slog.Logger
}

// new service
func NewHealthService(log *slog.Logger) *HealthService {
	return &HealthService{
		log: log.With("service", "health"),
	}
}
