// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/daytonaio/runner/internal/constants"
	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/netrules"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type DockerClientConfig struct {
	ApiClient              client.APIClient
	StatesCache            *cache.StatesCache
	LogWriter              io.Writer
	AWSRegion              string
	AWSEndpointUrl         string
	AWSAccessKeyId         string
	AWSSecretAccessKey     string
	DaemonPath             string
	ComputerUsePluginPath  string
	NetRulesManager        *netrules.NetRulesManager
	ResourceLimitsDisabled bool
}

func NewDockerClient(config DockerClientConfig) *DockerClient {
	return &DockerClient{
		apiClient:              config.ApiClient,
		statesCache:            config.StatesCache,
		logWriter:              config.LogWriter,
		awsRegion:              config.AWSRegion,
		awsEndpointUrl:         config.AWSEndpointUrl,
		awsAccessKeyId:         config.AWSAccessKeyId,
		awsSecretAccessKey:     config.AWSSecretAccessKey,
		volumeMutexes:          make(map[string]*sync.Mutex),
		daemonPath:             config.DaemonPath,
		computerUsePluginPath:  config.ComputerUsePluginPath,
		netRulesManager:        config.NetRulesManager,
		resourceLimitsDisabled: config.ResourceLimitsDisabled,
	}
}

func (d *DockerClient) ApiClient() client.APIClient {
	return d.apiClient
}

type DockerClient struct {
	apiClient              client.APIClient
	statesCache            *cache.StatesCache
	logWriter              io.Writer
	awsRegion              string
	awsEndpointUrl         string
	awsAccessKeyId         string
	awsSecretAccessKey     string
	volumeMutexes          map[string]*sync.Mutex
	volumeMutexesMutex     sync.Mutex
	daemonPath             string
	computerUsePluginPath  string
	netRulesManager        *netrules.NetRulesManager
	resourceLimitsDisabled bool
}

// retryWithExponentialBackoff executes a function with exponential backoff retry logic
func (d *DockerClient) retryWithExponentialBackoff(ctx context.Context, operationName, containerId string, maxRetries int, baseDelay, maxDelay time.Duration, operationFunc func() error) error {
	if maxRetries <= 1 {
		log.Debugf("Invalid max retries value: %d. Using default value: %d", maxRetries, constants.DEFAULT_MAX_RETRIES)
		maxRetries = constants.DEFAULT_MAX_RETRIES
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		logAttempt := attempt + 1
		log.Debugf("%s sandbox %s (attempt %d/%d)...", operationName, containerId, logAttempt, maxRetries)

		err := operationFunc()
		if err == nil {
			return nil
		}

		if attempt < maxRetries-1 {
			// Calculate exponential backoff delay
			delay := min(baseDelay*time.Duration(1<<attempt), maxDelay)

			log.Warnf("Failed to %s sandbox %s (attempt %d/%d): %v. Retrying in %v...", operationName, containerId, logAttempt, maxRetries, err, delay)

			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return fmt.Errorf("failed to %s sandbox after %d attempts: %w", operationName, maxRetries, err)
	}

	return nil
}
