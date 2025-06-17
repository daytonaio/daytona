// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"io"
	"log/slog"
	"sync"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/pkg/cache"
	"github.com/daytonaio/runner-docker/pkg/grpc/services/snapshot"
	"github.com/docker/docker/client"
)

type SandboxServiceConfig struct {
	DockerClient       client.APIClient
	SnapshotService    *snapshot.SnapshotService
	Cache              cache.IRunnerCache
	LogWriter          io.Writer
	DaemonPath         string
	AWSAccessKeyId     string
	AWSSecretAccessKey string
	AWSRegion          string
	AWSEndpointUrl     string
	ContainerNetwork   string
	ContainerRuntime   string
	NodeEnv            string
	Log                *slog.Logger
}

type SandboxService struct {
	pb.UnimplementedSandboxServiceServer
	snapshotService    *snapshot.SnapshotService
	dockerClient       client.APIClient
	cache              cache.IRunnerCache
	logWriter          io.Writer
	daemonPath         string
	awsEndpointUrl     string
	awsAccessKeyId     string
	awsSecretAccessKey string
	awsRegion          string
	containerNetwork   string
	containerRuntime   string
	nodeEnv            string
	volumeMutexes      map[string]*sync.Mutex
	volumeMutexesMutex sync.Mutex
	log                *slog.Logger
}

func NewSandboxService(config SandboxServiceConfig) *SandboxService {
	return &SandboxService{
		dockerClient:       config.DockerClient,
		snapshotService:    config.SnapshotService,
		cache:              config.Cache,
		logWriter:          config.LogWriter,
		daemonPath:         config.DaemonPath,
		awsEndpointUrl:     config.AWSEndpointUrl,
		awsAccessKeyId:     config.AWSAccessKeyId,
		awsSecretAccessKey: config.AWSSecretAccessKey,
		awsRegion:          config.AWSRegion,
		containerNetwork:   config.ContainerNetwork,
		containerRuntime:   config.ContainerRuntime,
		nodeEnv:            config.NodeEnv,
		volumeMutexes:      make(map[string]*sync.Mutex),
		volumeMutexesMutex: sync.Mutex{},
		log:                config.Log.With("service", "sandbox"),
	}
}
