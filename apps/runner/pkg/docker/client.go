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
	ApiClient          client.APIClient
	Cache              cache.IRunnerCache
	DaytonaBinaryUrl   string
	TerminalBinaryUrl  string
	DaytonaBinaryPath  string
	TerminalBinaryPath string
	LogWriter          io.Writer
	AWSRegion          string
	AWSEndpointUrl     string
	AWSAccessKeyId     string
	AWSSecretAccessKey string
}

func NewDockerClient(config DockerClientConfig) *DockerClient {
	return &DockerClient{
		apiClient:          config.ApiClient,
		cache:              config.Cache,
		daytonaBinaryURL:   config.DaytonaBinaryUrl,
		daytonaBinaryPath:  config.DaytonaBinaryPath,
		terminalBinaryURL:  config.TerminalBinaryUrl,
		terminalBinaryPath: config.TerminalBinaryPath,
		logWriter:          config.LogWriter,
		awsRegion:          config.AWSRegion,
		awsEndpointUrl:     config.AWSEndpointUrl,
		awsAccessKeyId:     config.AWSAccessKeyId,
		awsSecretAccessKey: config.AWSSecretAccessKey,
		volumeMutexes:      make(map[string]*sync.Mutex),
	}
}

type DockerClient struct {
	apiClient          client.APIClient
	cache              cache.IRunnerCache
	daytonaBinaryPath  string
	daytonaBinaryURL   string
	terminalBinaryPath string
	terminalBinaryURL  string
	logWriter          io.Writer
	awsRegion          string
	awsEndpointUrl     string
	awsAccessKeyId     string
	awsSecretAccessKey string
	volumeMutexes      map[string]*sync.Mutex
	volumeMutexesMutex sync.Mutex
}
