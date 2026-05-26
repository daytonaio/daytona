// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/daytonaio/common-go/pkg/utils"
	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/netrules"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
)

type DockerClientConfig struct {
	ApiClient                    client.APIClient
	BackupInfoCache              *cache.BackupInfoCache
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
	AndroidBootTimeoutSec        int
	UseSnapshotEntrypoint        bool
	VolumeCleanupInterval        time.Duration
	VolumeCleanupDryRun          bool
	VolumeCleanupExclusionPeriod time.Duration
	BackupTimeoutMin             int
	SnapshotPullTimeout          time.Duration
	BuildTimeoutMin              int
	BuildCPUCores                int64
	BuildMemoryGB                int64
	InitializeDaemonTelemetry    bool
	InterSandboxNetworkEnabled   bool
	GpuEnabled                   bool
	MountKvmToAndroidSandbox     bool
}

func NewDockerClient(ctx context.Context, config DockerClientConfig) (*DockerClient, error) {
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

	// Android emulator cold boot can take well over a minute even on capable hosts,
	// so we allow a dedicated longer timeout for the ADB readiness probe.
	if config.AndroidBootTimeoutSec <= 0 {
		logger.Warn("Invalid android boot timeout value. Using default value of 300 seconds")
		config.AndroidBootTimeoutSec = 300
	}

	if config.BackupTimeoutMin <= 0 {
		logger.Warn("Invalid backup timeout value. Using default value of 60 minutes")
		config.BackupTimeoutMin = 60
	}

	var info system.Info
	err := utils.RetryWithExponentialBackoff(
		ctx,
		"get Docker info",
		8,
		1*time.Second,
		5*time.Second,
		func() error {
			var infoErr error
			info, infoErr = config.ApiClient.Info(ctx)
			return infoErr
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker info: %w", err)
	}

	if !config.InterSandboxNetworkEnabled {
		if _, err := config.ApiClient.NetworkInspect(ctx, RUNNER_BRIDGE_NETWORK_NAME, network.InspectOptions{}); err != nil {
			_, err := config.ApiClient.NetworkCreate(ctx, RUNNER_BRIDGE_NETWORK_NAME, network.CreateOptions{
				Driver: "bridge",
				Options: map[string]string{
					"com.docker.network.bridge.enable_icc": "false",
				},
				IPAM: &network.IPAM{
					Driver: "default",
					Config: []network.IPAMConfig{
						{Subnet: "172.20.0.0/16"},
					},
				},
			})

			if err != nil {
				return nil, fmt.Errorf("failed to create %s network: %w", RUNNER_BRIDGE_NETWORK_NAME, err)
			}
		}
	}

	filesystem := ""

	for _, driver := range info.DriverStatus {
		if driver[0] == "Backing Filesystem" {
			filesystem = driver[1]
			break
		}
	}

	gpuCount := 0
	gpuType := ""
	if config.GpuEnabled {
		gpuCount, gpuType = detectGpus(ctx)
		if gpuCount == 0 {
			logger.Warn("GPU_ENABLED=true but nvidia-smi did not report any GPUs; runner will not host GPU sandboxes")
		} else {
			logger.Info("Detected GPUs", "count", gpuCount, "type", gpuType)
		}
	}

	return &DockerClient{
		apiClient:                    config.ApiClient,
		backupInfoCache:              config.BackupInfoCache,
		pullTracker:                  &common.Tracker[string]{},
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
		androidBootTimeoutSec:        config.AndroidBootTimeoutSec,
		useSnapshotEntrypoint:        config.UseSnapshotEntrypoint,
		volumeCleanupInterval:        config.VolumeCleanupInterval,
		volumeCleanupDryRun:          config.VolumeCleanupDryRun,
		volumeCleanupExclusionPeriod: config.VolumeCleanupExclusionPeriod,
		backupTimeoutMin:             config.BackupTimeoutMin,
		snapshotPullTimeout:          config.SnapshotPullTimeout,
		buildTimeoutMin:              config.BuildTimeoutMin,
		buildCPUCores:                config.BuildCPUCores,
		buildMemoryGB:                config.BuildMemoryGB,
		initializeDaemonTelemetry:    config.InitializeDaemonTelemetry,
		interSandboxNetworkEnabled:   config.InterSandboxNetworkEnabled,
		gpuEnabled:                   config.GpuEnabled,
		gpuCount:                     gpuCount,
		gpuType:                      gpuType,
		gpuAllocator:                 newGpuAllocator(gpuCount),
		filesystem:                   filesystem,
		mountKvmToAndroidSandbox:     config.MountKvmToAndroidSandbox,
	}, nil
}

// GpuCount returns the number of NVIDIA GPUs detected on the host at startup.
// Returns 0 when GPU support is disabled or no GPU is present.
func (d *DockerClient) GpuCount() int {
	return d.gpuCount
}

// GpuType returns the human-readable name of the first GPU detected on the
// host at startup (e.g. "NVIDIA H100 80GB HBM3"). Returns "" when no GPU is
// present.
func (d *DockerClient) GpuType() string {
	return d.gpuType
}

func (d *DockerClient) ApiClient() client.APIClient {
	return d.apiClient
}

const RUNNER_BRIDGE_NETWORK_NAME = "runner-bridge"

type DockerClient struct {
	apiClient                    client.APIClient
	backupInfoCache              *cache.BackupInfoCache
	pullTracker                  *common.Tracker[string]
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
	androidBootTimeoutSec        int
	useSnapshotEntrypoint        bool
	volumeCleanupInterval        time.Duration
	volumeCleanupDryRun          bool
	volumeCleanupExclusionPeriod time.Duration
	backupTimeoutMin             int
	snapshotPullTimeout          time.Duration
	buildTimeoutMin              int
	buildCPUCores                int64
	buildMemoryGB                int64
	volumeCleanupMutex           sync.Mutex
	lastVolumeCleanup            time.Time
	initializeDaemonTelemetry    bool
	filesystem                   string
	interSandboxNetworkEnabled   bool
	gpuEnabled                   bool
	gpuCount                     int
	gpuType                      string
	gpuAllocator                 *gpuAllocator
	mountKvmToAndroidSandbox     bool
}
