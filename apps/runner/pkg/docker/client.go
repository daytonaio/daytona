// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"io"
	"sync"

	"github.com/daytonaio/runner/pkg/cache"
	"github.com/docker/docker/client"
)

type DockerClientConfig struct {
	ApiClient              client.APIClient
	Cache                  cache.IRunnerCache
	LogWriter              io.Writer
	AWSRegion              string
	AWSEndpointUrl         string
	AWSAccessKeyId         string
	AWSSecretAccessKey     string
	DaemonPath             string
	ComputerUsePluginPath  string
	ResourceLimitsDisabled bool
}

func NewDockerClient(config DockerClientConfig) *DockerClient {
	return &DockerClient{
		apiClient:              config.ApiClient,
		cache:                  config.Cache,
		logWriter:              config.LogWriter,
		awsRegion:              config.AWSRegion,
		awsEndpointUrl:         config.AWSEndpointUrl,
		awsAccessKeyId:         config.AWSAccessKeyId,
		awsSecretAccessKey:     config.AWSSecretAccessKey,
		volumeMutexes:          make(map[string]*sync.Mutex),
		daemonPath:             config.DaemonPath,
		computerUsePluginPath:  config.ComputerUsePluginPath,
		resourceLimitsDisabled: config.ResourceLimitsDisabled,
	}
}

func (d *DockerClient) ApiClient() client.APIClient {
	return d.apiClient
}

type DockerClient struct {
	apiClient              client.APIClient
	cache                  cache.IRunnerCache
	logWriter              io.Writer
	awsRegion              string
	awsEndpointUrl         string
	awsAccessKeyId         string
	awsSecretAccessKey     string
	volumeMutexes          map[string]*sync.Mutex
	volumeMutexesMutex     sync.Mutex
	daemonPath             string
	computerUsePluginPath  string
	resourceLimitsDisabled bool
}
