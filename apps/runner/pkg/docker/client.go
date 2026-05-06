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
	"github.com/daytonaio/runner/pkg/volume"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
)

type DockerClientConfig struct {
	ApiClient                    client.APIClient
	BackupInfoCache              *cache.BackupInfoCache
	Logger                       *slog.Logger
	DefaultVolumeMounter         volume.Mounter
	InContainerVolumeMounter     volume.Mounter // optional; when nil, the "experimental" backend silently falls back to s3fuse
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

	return &DockerClient{
		apiClient:                    config.ApiClient,
		backupInfoCache:              config.BackupInfoCache,
		pullTracker:                  &common.Tracker[string]{},
		logger:                       logger,
		defaultVolumeMounter:         config.DefaultVolumeMounter,
		inContainerVolumeMounter:     config.InContainerVolumeMounter,
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
		filesystem:                   filesystem,
	}, nil
}

func (d *DockerClient) ApiClient() client.APIClient {
	return d.apiClient
}

const volumeBackendMetadataKey = "volumeBackend"

// Volume backend identifiers exchanged with the control plane via sandbox metadata.
const (
	// volumeBackendS3Fuse is the default: the runner mounts the S3 bucket on
	// the host using its own AWS credentials and bind-mounts the host
	// mountpoint into the sandbox. Used when no backend is explicitly
	// requested or for any unknown backend value.
	volumeBackendS3Fuse = "s3fuse"

	// volumeBackendExperimental routes to the in-container mounter, which
	// mounts an Archil disk from inside the sandbox using a per-volume
	// ARCHIL_MOUNT_TOKEN. Falls back to s3fuse if the runner has no
	// in-container mounter configured.
	volumeBackendExperimental = "experimental"
)

// resolveVolumeMounter selects the volume mounter based on the per-sandbox
// metadata key. "experimental" routes to the in-container (Archil) mounter
// when configured; everything else falls back to the host-side s3fuse default.
func (d *DockerClient) resolveVolumeMounter(metadata map[string]string) volume.Mounter {
	if metadata[volumeBackendMetadataKey] == volumeBackendExperimental && d.inContainerVolumeMounter != nil {
		return d.inContainerVolumeMounter
	}
	return d.defaultVolumeMounter
}

const RUNNER_BRIDGE_NETWORK_NAME = "runner-bridge"

type DockerClient struct {
	apiClient                    client.APIClient
	backupInfoCache              *cache.BackupInfoCache
	pullTracker                  *common.Tracker[string]
	logger                       *slog.Logger
	defaultVolumeMounter         volume.Mounter
	inContainerVolumeMounter     volume.Mounter // nil when the experimental backend is not configured
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
}
