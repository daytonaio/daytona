// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"io"
	"log/slog"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/pkg/cache"
	"github.com/docker/docker/client"
)

type SnapshotServiceConfig struct {
	DockerClient client.APIClient
	Cache        cache.IRunnerCache
	LogWriter    io.Writer
	Log          *slog.Logger
	LogFilePath  string
}

type SnapshotService struct {
	pb.UnimplementedSnapshotServiceServer
	dockerClient client.APIClient
	cache        cache.IRunnerCache
	logWriter    io.Writer
	log          *slog.Logger
	logFilePath  string
}

func NewSnapshotService(config SnapshotServiceConfig) *SnapshotService {
	return &SnapshotService{
		dockerClient: config.DockerClient,
		cache:        config.Cache,
		logWriter:    config.LogWriter,
		log:          config.Log.With("service", "snapshot"),
		logFilePath:  config.LogFilePath,
	}
}
