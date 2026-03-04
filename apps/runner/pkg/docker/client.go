// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/netrules"
	"github.com/docker/docker/client"
)

type DockerClientConfig struct {
	ApiClient                    client.APIClient
	StatesCache                  *cache.StatesCache
	Logger                       *slog.Logger
	AWSRegion                    string
	AWSEndpointUrl               string
	AWSAccessKeyId               string
	AWSSecretAccessKey           string
	DaemonPath                   string
	ComputerUsePluginPath        string
	NetRulesManager              *netrules.NetRulesManager
	ResourceLimitsDisabled       bool
	DaemonStartTimeoutSec        int
	SandboxStartTimeoutSec       int
	UseSnapshotEntrypoint        bool
	VolumeCleanupInterval        time.Duration
	VolumeCleanupDryRun          bool
	VolumeCleanupExclusionPeriod time.Duration
	BackupTimeoutMin             int
	InitializeDaemonTelemetry    bool
	VolumeMountDriver            string
}

func NewDockerClient(config DockerClientConfig) (*DockerClient, error) {
	logger := slog.Default().With(slog.String("component", "docker-client"))
	if config.Logger != nil {
		logger = config.Logger.With(slog.String("component", "docker-client"))
	}

	if config.DaemonStartTimeoutSec <= 0 {
		logger.Warn("Invalid daemon start timeout value. Using default value of 60 seconds")
		config.DaemonStartTimeoutSec = 60
	}

	if config.SandboxStartTimeoutSec <= 0 {
		logger.Warn("Invalid sandbox start timeout value. Using default value of 30 seconds")
		config.SandboxStartTimeoutSec = 30
	}

	if config.BackupTimeoutMin <= 0 {
		logger.Warn("Invalid backup timeout value. Using default value of 60 minutes")
		config.BackupTimeoutMin = 60
	}

	info, err := config.ApiClient.Info(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker info: %w", err)
	}

	filesystem := ""

	for _, driver := range info.DriverStatus {
		if driver[0] == "Backing Filesystem" {
			filesystem = driver[1]
			break
		}
	}

	return &DockerClient{
		apiClient:                    config.ApiClient,
		statesCache:                  config.StatesCache,
		logger:                       logger,
		awsRegion:                    config.AWSRegion,
		awsEndpointUrl:               config.AWSEndpointUrl,
		awsAccessKeyId:               config.AWSAccessKeyId,
		awsSecretAccessKey:           config.AWSSecretAccessKey,
		volumeMutexes:                make(map[string]*sync.Mutex),
		daemonPath:                   config.DaemonPath,
		computerUsePluginPath:        config.ComputerUsePluginPath,
		netRulesManager:              config.NetRulesManager,
		resourceLimitsDisabled:       config.ResourceLimitsDisabled,
		daemonStartTimeoutSec:        config.DaemonStartTimeoutSec,
		sandboxStartTimeoutSec:       config.SandboxStartTimeoutSec,
		useSnapshotEntrypoint:        config.UseSnapshotEntrypoint,
		volumeCleanupInterval:        config.VolumeCleanupInterval,
		volumeCleanupDryRun:          config.VolumeCleanupDryRun,
		volumeCleanupExclusionPeriod: config.VolumeCleanupExclusionPeriod,
		backupTimeoutMin:             config.BackupTimeoutMin,
		initializeDaemonTelemetry:    config.InitializeDaemonTelemetry,
		filesystem:                   filesystem,
		volumeMountDriver:            config.VolumeMountDriver,
	}, nil
}

func (d *DockerClient) ApiClient() client.APIClient {
	return d.apiClient
}

type DockerClient struct {
	apiClient                    client.APIClient
	statesCache                  *cache.StatesCache
	logger                       *slog.Logger
	awsRegion                    string
	awsEndpointUrl               string
	awsAccessKeyId               string
	awsSecretAccessKey           string
	volumeMutexes                map[string]*sync.Mutex
	volumeMutexesMutex           sync.Mutex
	daemonPath                   string
	computerUsePluginPath        string
	netRulesManager              *netrules.NetRulesManager
	resourceLimitsDisabled       bool
	daemonStartTimeoutSec        int
	sandboxStartTimeoutSec       int
	useSnapshotEntrypoint        bool
	volumeCleanupInterval        time.Duration
	volumeCleanupDryRun          bool
	volumeCleanupExclusionPeriod time.Duration
	backupTimeoutMin             int
	volumeCleanupMutex           sync.Mutex
	lastVolumeCleanup            time.Time
	initializeDaemonTelemetry    bool
	filesystem                   string
	volumeMountDriver            string
}
