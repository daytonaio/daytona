// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"io"

	"github.com/daytonaio/runner/pkg/cache"
	"github.com/docker/docker/client"
)

type DockerClientConfig struct {
	ApiClient          client.APIClient
	Cache              cache.IRunnerCache
	LogWriter          io.Writer
	AWSRegion          string
	AWSEndpointUrl     string
	AWSAccessKeyId     string
	AWSSecretAccessKey string
	DaemonPath         string
}

func NewDockerClient(config DockerClientConfig) *DockerClient {
	return &DockerClient{
		apiClient:          config.ApiClient,
		cache:              config.Cache,
		logWriter:          config.LogWriter,
		awsRegion:          config.AWSRegion,
		awsEndpointUrl:     config.AWSEndpointUrl,
		awsAccessKeyId:     config.AWSAccessKeyId,
		awsSecretAccessKey: config.AWSSecretAccessKey,
		daemonPath:         config.DaemonPath,
	}
}

type DockerClient struct {
	apiClient          client.APIClient
	cache              cache.IRunnerCache
	logWriter          io.Writer
	awsRegion          string
	awsEndpointUrl     string
	awsAccessKeyId     string
	awsSecretAccessKey string
	daemonPath         string
}
