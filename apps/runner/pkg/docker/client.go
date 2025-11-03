// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/daytonaio/runner/internal/constants"
	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/netrules"
	"github.com/docker/docker/client"
)

type DockerClientConfig struct {
	ApiClient              client.APIClient
	StatesCache            *cache.StatesCache
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
		slog.DebugContext(ctx, "Invalid max retries value. Using default value", "maxRetries", maxRetries, "defaultMaxRetries", constants.DEFAULT_MAX_RETRIES)
		maxRetries = constants.DEFAULT_MAX_RETRIES
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		logAttempt := attempt + 1
		slog.DebugContext(ctx, "Attempting operation", "operationName", operationName, "containerId", containerId, "attempt", logAttempt, "maxRetries", maxRetries)

		err := operationFunc()
		if err == nil {
			return nil
		}

		if attempt < maxRetries-1 {
			// Calculate exponential backoff delay
			delay := baseDelay * time.Duration(1<<(attempt-1))
			if delay > maxDelay {
				delay = maxDelay
			}

			slog.WarnContext(ctx, "Operation failed, retrying", "operationName", operationName, "containerId", containerId, "attempt", logAttempt, "maxRetries", maxRetries, "error", err, "delay", delay)

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
