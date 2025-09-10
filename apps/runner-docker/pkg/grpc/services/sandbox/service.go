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
	"github.com/daytonaio/runner-docker/pkg/netrules"
	"github.com/docker/docker/client"
)

type SandboxServiceConfig struct {
	DockerClient          client.APIClient
	SnapshotService       *snapshot.SnapshotService
	Cache                 cache.IRunnerCache
	LogWriter             io.Writer
	DaemonPath            string
	AWSAccessKeyId        string
	AWSSecretAccessKey    string
	AWSRegion             string
	AWSEndpointUrl        string
	ContainerNetwork      string
	ContainerRuntime      string
	Environment           string
	Log                   *slog.Logger
	ComputerUsePluginPath string
	NetRulesManager       *netrules.NetRulesManager
}

type SandboxService struct {
	pb.UnimplementedSandboxServiceServer
	snapshotService       *snapshot.SnapshotService
	dockerClient          client.APIClient
	cache                 cache.IRunnerCache
	logWriter             io.Writer
	daemonPath            string
	awsEndpointUrl        string
	awsAccessKeyId        string
	awsSecretAccessKey    string
	awsRegion             string
	containerNetwork      string
	containerRuntime      string
	environment           string
	volumeMutexes         map[string]*sync.Mutex
	volumeMutexesMutex    sync.Mutex
	log                   *slog.Logger
	computerUsePluginPath string
	netRulesManager       *netrules.NetRulesManager
}

func NewSandboxService(config SandboxServiceConfig) *SandboxService {
	return &SandboxService{
		dockerClient:          config.DockerClient,
		snapshotService:       config.SnapshotService,
		cache:                 config.Cache,
		logWriter:             config.LogWriter,
		daemonPath:            config.DaemonPath,
		awsEndpointUrl:        config.AWSEndpointUrl,
		awsAccessKeyId:        config.AWSAccessKeyId,
		awsSecretAccessKey:    config.AWSSecretAccessKey,
		awsRegion:             config.AWSRegion,
		containerNetwork:      config.ContainerNetwork,
		containerRuntime:      config.ContainerRuntime,
		environment:           config.Environment,
		volumeMutexes:         make(map[string]*sync.Mutex),
		volumeMutexesMutex:    sync.Mutex{},
		log:                   config.Log.With("service", "sandbox"),
		computerUsePluginPath: config.ComputerUsePluginPath,
		netRulesManager:       config.NetRulesManager,
	}
}
