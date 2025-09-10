// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package runner

import (
	"log/slog"
	"time"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/pkg/cache"
	"github.com/docker/docker/client"
)

type RunnerServiceConfig struct {
	Log          *slog.Logger
	DockerClient client.APIClient
	Cache        cache.ICache[SystemMetrics]
	Interval     time.Duration
}

type RunnerService struct {
	pb.UnimplementedRunnerServiceServer
	log          *slog.Logger
	dockerClient client.APIClient
	cache        cache.ICache[SystemMetrics]
	interval     time.Duration
}

// new service
func NewRunnerService(config RunnerServiceConfig) *RunnerService {
	return &RunnerService{
		log:          config.Log.With("service", "runner"),
		dockerClient: config.DockerClient,
		cache:        config.Cache,
		interval:     config.Interval,
	}
}
