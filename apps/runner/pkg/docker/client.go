// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"io"
	"sync"
	"time"

	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/netrules"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type DockerClientConfig struct {
	ApiClient                    client.APIClient
	StatesCache                  *cache.StatesCache
	LogWriter                    io.Writer
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
}

func NewDockerClient(config DockerClientConfig) *DockerClient {
	if config.DaemonStartTimeoutSec <= 0 {
		log.Warnf("Invalid DaemonStartTimeoutSec value: %d. Using default value: 60 seconds", config.DaemonStartTimeoutSec)
		config.DaemonStartTimeoutSec = 60
	}

	if config.SandboxStartTimeoutSec <= 0 {
		log.Warnf("Invalid SandboxStartTimeoutSec value: %d. Using default value: 30 seconds", config.SandboxStartTimeoutSec)
		config.SandboxStartTimeoutSec = 30
	}

	if config.BackupTimeoutMin <= 0 {
		log.Warnf("Invalid BackupTimeoutMin value: %d. Using default value: 60 minutes", config.BackupTimeoutMin)
		config.BackupTimeoutMin = 60
	}

	return &DockerClient{
		apiClient:                     config.ApiClient,
		statesCache:                   config.StatesCache,
		logWriter:                     config.LogWriter,
		awsRegion:                     config.AWSRegion,
		awsEndpointUrl:                config.AWSEndpointUrl,
		awsAccessKeyId:                config.AWSAccessKeyId,
		awsSecretAccessKey:            config.AWSSecretAccessKey,
		volumeMutexes:                 make(map[string]*sync.Mutex),
		daemonPath:                    config.DaemonPath,
		computerUsePluginPath:         config.ComputerUsePluginPath,
		netRulesManager:               config.NetRulesManager,
		resourceLimitsDisabled:        config.ResourceLimitsDisabled,
		daemonStartTimeoutSec:         config.DaemonStartTimeoutSec,
		sandboxStartTimeoutSec:        config.SandboxStartTimeoutSec,
		useSnapshotEntrypoint:         config.UseSnapshotEntrypoint,
		volumeCleanupInterval:         config.VolumeCleanupInterval,
		volumeCleanupDryRun:           config.VolumeCleanupDryRun,
		volumeCleanupExclusionPeriod:  config.VolumeCleanupExclusionPeriod,
		backupTimeoutMin:              config.BackupTimeoutMin,
		initializeDaemonTelemetry:     config.InitializeDaemonTelemetry,
	}
}

func (d *DockerClient) ApiClient() client.APIClient {
	return d.apiClient
}

type DockerClient struct {
	apiClient                    client.APIClient
	statesCache                  *cache.StatesCache
	logWriter                    io.Writer
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
}
