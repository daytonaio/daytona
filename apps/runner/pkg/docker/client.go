// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/netrules"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type DockerClientConfig struct {
	ApiClient             client.APIClient
	Cache                 cache.IRunnerCache
	LogWriter             io.Writer
	AWSRegion             string
	AWSEndpointUrl        string
	AWSAccessKeyId        string
	AWSSecretAccessKey    string
	DaemonPath            string
	ComputerUsePluginPath string
	NetRulesManager       *netrules.NetRulesManager
}

func NewDockerClient(config DockerClientConfig) *DockerClient {
	return &DockerClient{
		apiClient:             config.ApiClient,
		cache:                 config.Cache,
		logWriter:             config.LogWriter,
		awsRegion:             config.AWSRegion,
		awsEndpointUrl:        config.AWSEndpointUrl,
		awsAccessKeyId:        config.AWSAccessKeyId,
		awsSecretAccessKey:    config.AWSSecretAccessKey,
		volumeMutexes:         make(map[string]*sync.Mutex),
		daemonPath:            config.DaemonPath,
		computerUsePluginPath: config.ComputerUsePluginPath,
		netRulesManager:       config.NetRulesManager,
	}
}

func (d *DockerClient) ApiClient() client.APIClient {
	return d.apiClient
}

func (d *DockerClient) Cache() cache.IRunnerCache {
	return d.cache
}

type DockerClient struct {
	apiClient             client.APIClient
	cache                 cache.IRunnerCache
	logWriter             io.Writer
	awsRegion             string
	awsEndpointUrl        string
	awsAccessKeyId        string
	awsSecretAccessKey    string
	volumeMutexes         map[string]*sync.Mutex
	volumeMutexesMutex    sync.Mutex
	daemonPath            string
	computerUsePluginPath string
	netRulesManager       *netrules.NetRulesManager
}

// retryWithExponentialBackoff executes a function with exponential backoff retry logic
func (d *DockerClient) retryWithExponentialBackoff(operationName, containerId string, maxRetries int, baseDelay, maxDelay time.Duration, operationFunc func() error) error {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Infof("%s container %s (attempt %d/%d)...", operationName, containerId, attempt, maxRetries)

		err := operationFunc()
		if err == nil {
			return nil
		}

		if attempt < maxRetries {
			// Calculate exponential backoff delay
			delay := baseDelay * time.Duration(1<<(attempt-1))
			if delay > maxDelay {
				delay = maxDelay
			}

			log.Warnf("Failed to %s container %s (attempt %d/%d): %v. Retrying in %v...", operationName, containerId, attempt, maxRetries, err, delay)
			time.Sleep(delay)
			continue
		}

		return fmt.Errorf("failed to %s container after %d attempts: %w", operationName, maxRetries, err)
	}

	return nil
}
