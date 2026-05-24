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
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
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
	var hostCPUCores, hostMemoryGiB, hostDiskGiB int64
	if config.GpuEnabled {
		gpuCount, gpuType = detectGpus(ctx)
		if gpuCount == 0 {
			logger.Warn("GPU_ENABLED=true but nvidia-smi did not report any GPUs; runner will not host GPU sandboxes")
		} else {
			logger.Info("Detected GPUs", "count", gpuCount, "type", gpuType)
			hostCPUCores, hostMemoryGiB, hostDiskGiB = detectHostCapacity(ctx, logger)
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
		hostCPUCores:                 hostCPUCores,
		hostMemoryGiB:                hostMemoryGiB,
		hostDiskGiB:                  hostDiskGiB,
		filesystem:                   filesystem,
	}, nil
}

// detectHostCapacity reads the host's total CPU cores, RAM, and the size of
// the filesystem backing /var/lib/docker. It is called once at startup, only
// when GPU support is enabled, so it can be used to compute the per-GPU slice
// for GPU sandboxes (host_total / gpu_count).
func detectHostCapacity(ctx context.Context, logger *slog.Logger) (int64, int64, int64) {
	var cores, memGiB, diskGiB int64

	if n, err := cpu.CountsWithContext(ctx, true); err == nil {
		cores = int64(n)
	} else {
		logger.Warn("Failed to detect host CPU count; GPU sandbox CPU slice unavailable", "error", err)
	}

	if m, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		memGiB = int64(m.Total / (1024 * 1024 * 1024))
	} else {
		logger.Warn("Failed to detect host memory; GPU sandbox memory slice unavailable", "error", err)
	}

	if d, err := disk.UsageWithContext(ctx, "/var/lib/docker"); err == nil {
		diskGiB = int64(d.Total / (1024 * 1024 * 1024))
	} else {
		logger.Warn("Failed to detect host disk; GPU sandbox disk slice unavailable", "error", err)
	}

	logger.Info("Detected host capacity", "cpuCores", cores, "memoryGiB", memGiB, "diskGiB", diskGiB)
	return cores, memGiB, diskGiB
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
	hostCPUCores                 int64
	hostMemoryGiB                int64
	hostDiskGiB                  int64
}
