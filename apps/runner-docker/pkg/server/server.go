// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"io"
	"sync"

	"github.com/daytonaio/runner-docker/pkg/cache"
	pb "github.com/daytonaio/runner/proto"
	"github.com/docker/docker/client"
)

type RunnerServerConfig struct {
	ApiClient          client.APIClient
	Cache              cache.IRunnerCache
	LogWriter          io.Writer
	AWSRegion          string
	AWSEndpointUrl     string
	AWSAccessKeyId     string
	AWSSecretAccessKey string
	DaemonPath         string
}

type RunnerServer struct {
	pb.UnimplementedRunnerServer
	apiClient          client.APIClient
	cache              cache.IRunnerCache
	logWriter          io.Writer
	awsRegion          string
	awsEndpointUrl     string
	awsAccessKeyId     string
	awsSecretAccessKey string
	daemonPath         string
	volumeMutexes      map[string]*sync.Mutex
	volumeMutexesMutex sync.Mutex
}

func NewRunnerServer(config RunnerServerConfig) *RunnerServer {
	return &RunnerServer{
		apiClient:          config.ApiClient,
		cache:              config.Cache,
		logWriter:          config.LogWriter,
		awsRegion:          config.AWSRegion,
		awsEndpointUrl:     config.AWSEndpointUrl,
		awsAccessKeyId:     config.AWSAccessKeyId,
		awsSecretAccessKey: config.AWSSecretAccessKey,
		daemonPath:         config.DaemonPath,
		volumeMutexes:      make(map[string]*sync.Mutex),
	}
}
