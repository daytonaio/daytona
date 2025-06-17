// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"io"
	"sync"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/pkg/cache"
	"github.com/docker/docker/client"
)

type SandboxServiceConfig struct {
	DockerClient       client.APIClient
	Cache              cache.IRunnerCache
	LogWriter          io.Writer
	DaemonPath         string
	AWSAccessKeyId     string
	AWSSecretAccessKey string
	AWSRegion          string
	AWSEndpointUrl     string
}

type SandboxService struct {
	pb.UnimplementedSandboxServiceServer
	dockerClient       client.APIClient
	cache              cache.IRunnerCache
	logWriter          io.Writer
	daemonPath         string
	awsEndpointUrl     string
	awsAccessKeyId     string
	awsSecretAccessKey string
	awsRegion          string
	volumeMutexes      map[string]*sync.Mutex
	volumeMutexesMutex sync.Mutex
}

func NewSandboxService(config SandboxServiceConfig) *SandboxService {
	return &SandboxService{
		dockerClient:       config.DockerClient,
		cache:              config.Cache,
		logWriter:          config.LogWriter,
		daemonPath:         config.DaemonPath,
		awsEndpointUrl:     config.AWSEndpointUrl,
		awsAccessKeyId:     config.AWSAccessKeyId,
		awsSecretAccessKey: config.AWSSecretAccessKey,
		awsRegion:          config.AWSRegion,
		volumeMutexes:      make(map[string]*sync.Mutex),
		volumeMutexesMutex: sync.Mutex{},
	}
}
