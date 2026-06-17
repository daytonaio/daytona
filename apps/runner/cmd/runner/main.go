// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daytonaio/common-go/pkg/log"
	"github.com/daytonaio/common-go/pkg/telemetry"
	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/internal"
	"github.com/daytonaio/runner/internal/metrics"
	"github.com/daytonaio/runner/pkg/api"
	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/daemon"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/netrules"
	"github.com/daytonaio/runner/pkg/runner"
	"github.com/daytonaio/runner/pkg/runner/v2/executor"
	"github.com/daytonaio/runner/pkg/runner/v2/healthcheck"
	"github.com/daytonaio/runner/pkg/runner/v2/poller"
	"github.com/daytonaio/runner/pkg/services"
	"github.com/daytonaio/runner/pkg/sshgateway"
	"github.com/daytonaio/runner/pkg/telemetry/filters"
	"github.com/daytonaio/runner/pkg/volume"
	"github.com/daytonaio/runner/pkg/volume/incontainer"
	"github.com/daytonaio/runner/pkg/volume/s3fuse"
	"github.com/docker/docker/client"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"go.opentelemetry.io/otel"
)

func main() {
	os.Exit(run())
}

func run() int {
	// Init slog logger
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
		TimeFormat: time.RFC3339,
		Level:      log.ParseLogLevel(os.Getenv("LOG_LEVEL")),
	}))

	slog.SetDefault(logger)

	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error("Failed to get config", "error", err)
		return 2
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.OtelLoggingEnabled && cfg.OtelEndpoint != "" {
		logger.Info("OpenTelemetry logging is enabled")

		telemetryConfig := telemetry.Config{
			Endpoint:       cfg.OtelEndpoint,
			Headers:        cfg.GetOtelHeaders(),
			ServiceName:    "daytona-runner",
			ServiceVersion: internal.Version,
			Environment:    cfg.Environment,
		}

		// Initialize OpenTelemetry logging
		newLogger, lp, err := telemetry.InitLogger(ctx, logger, telemetryConfig)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to initialize logger", "error", err)
			return 2
		}

		// Reassign logger to the new OTEL-enabled logger returned by InitLogger.
		// This ensures that all subsequent code uses the logger instance that has OTEL support.
		logger = newLogger

		defer telemetry.ShutdownLogger(logger, lp)
	}

	if cfg.OtelTracingEnabled && cfg.OtelEndpoint != "" {
		logger.Info("OpenTelemetry tracing is enabled")

		telemetryConfig := telemetry.Config{
			Endpoint:       cfg.OtelEndpoint,
			Headers:        cfg.GetOtelHeaders(),
			ServiceName:    "daytona-runner",
			ServiceVersion: internal.Version,
			Environment:    cfg.Environment,
		}

		// Initialize OpenTelemetry tracing with a custom filter to ignore 404 errors
		tp, err := telemetry.InitTracer(ctx, telemetryConfig, &filters.NotFoundExporterFilter{})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to initialize tracer", "error", err)
			return 2
		}
		defer telemetry.ShutdownTracer(logger, tp)
	}

	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
		client.WithTraceProvider(otel.GetTracerProvider()),
	)
	if err != nil {
		logger.Error("Error creating Docker client", "error", err)
		return 2
	}

	// Initialize net rules manager
	persistent := cfg.Environment != "development"
	netRulesManager, err := netrules.NewNetRulesManager(logger, persistent)
	if err != nil {
		logger.Error("Failed to initialize net rules manager", "error", err)
		return 2
	}

	// Start net rules manager
	if err = netRulesManager.Start(); err != nil {
		logger.Error("Failed to start net rules manager", "error", err)
		return 2
	}
	defer netRulesManager.Stop()

	daemonPath, err := daemon.WriteStaticBinary("daemon-amd64")
	if err != nil {
		logger.Error("Error writing daemon binary", "error", err)
		return 2
	}

	pluginPath, err := daemon.WriteStaticBinary("daytona-computer-use")
	if err != nil {
		logger.Error("Error writing plugin binary", "error", err)
		return 2
	}

	backupInfoCache := cache.NewBackupInfoCache(ctx, cfg.BackupInfoCacheRetention)

	defaultVolumeMounter := volume.Mounter(s3fuse.NewMounter(s3fuse.Config{
		AWSRegion:          cfg.AWSRegion,
		AWSEndpointUrl:     cfg.AWSEndpointUrl,
		AWSAccessKeyId:     cfg.AWSAccessKeyId,
		AWSSecretAccessKey: cfg.AWSSecretAccessKey,
	}, logger))

	// Layered in-container mounter, registered only when LAYERED_BINARY_PATH
	// points at the installed CLI; otherwise the "layered" backend is
	// disabled and falls back to host-side "s3fuse". Mount tokens come from
	// the control plane per volume, so the runner needs no layered API key.
	inContainerVolumeMounter, err := maybeBuildInContainerMounter(ctx, cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize layered in-container volume backend", "error", err)
		return 2
	}

	dockerClient, err := docker.NewDockerClient(ctx, docker.DockerClientConfig{
		ApiClient:                    cli,
		BackupInfoCache:              backupInfoCache,
		Logger:                       logger,
		DefaultVolumeMounter:         defaultVolumeMounter,
		InContainerVolumeMounter:     inContainerVolumeMounter,
		DaemonPath:                   daemonPath,
		ComputerUsePluginPath:        pluginPath,
		NetRulesManager:              netRulesManager,
		ResourceLimitsDisabled:       cfg.ResourceLimitsDisabled,
		DaemonStartTimeoutSec:        cfg.DaemonStartTimeoutSec,
		SandboxStartTimeoutSec:       cfg.SandboxStartTimeoutSec,
		AndroidBootTimeoutSec:        cfg.AndroidBootTimeoutSec,
		UseSnapshotEntrypoint:        cfg.UseSnapshotEntrypoint,
		VolumeCleanupInterval:        cfg.VolumeCleanupInterval,
		VolumeCleanupDryRun:          cfg.VolumeCleanupDryRun,
		VolumeCleanupExclusionPeriod: cfg.VolumeCleanupExclusionPeriod,
		BackupTimeoutMin:             cfg.BackupTimeoutMin,
		SnapshotPullTimeout:          cfg.SnapshotPullTimeout,
		BuildTimeoutMin:              cfg.BuildTimeoutMin,
		BuildCPUCores:                cfg.BuildCPUCores,
		BuildMemoryGB:                cfg.BuildMemoryGB,
		InitializeDaemonTelemetry:    cfg.InitializeDaemonTelemetry,
		InterSandboxNetworkEnabled:   cfg.InterSandboxNetworkEnabled,
		GpuEnabled:                   cfg.GpuEnabled,
		MountKvmToAndroidSandbox:     cfg.MountKvmToAndroidSandbox,
	})
	if err != nil {
		logger.Error("Error creating Docker client wrapper", "error", err)
		return 2
	}

	// Start Docker events monitor
	monitorOpts := docker.MonitorOptions{
		OnDestroyEvent: func(ctx context.Context) {
			dockerClient.CleanupOrphanedVolumeMounts(ctx)
		},
	}
	monitor := docker.NewDockerMonitor(logger, cli, netRulesManager, monitorOpts)
	monitorErrChan := make(chan error)
	go func() {
		logger.Info("Starting Docker monitor")
		err = monitor.Start()
		if err != nil {
			monitorErrChan <- err
		}
	}()
	defer monitor.Stop()

	sandboxService := services.NewSandboxService(logger, backupInfoCache, dockerClient)

	// Initialize sandbox state synchronization service
	sandboxSyncService := services.NewSandboxSyncService(services.SandboxSyncServiceConfig{
		Logger:   logger,
		Docker:   dockerClient,
		Interval: 10 * time.Second, // Sync every 10 seconds
	})
	sandboxSyncService.StartSyncProcess(ctx)

	// Initialize SSH Gateway if enabled
	var sshGatewayService *sshgateway.Service
	if sshgateway.IsSSHGatewayEnabled() {
		sshGatewayService = sshgateway.NewService(logger, dockerClient)

		go func() {
			logger.Info("Starting SSH Gateway")
			if err := sshGatewayService.Start(ctx); err != nil {
				logger.Error("SSH Gateway error", "error", err)
			}
		}()
	} else {
		logger.Info("Gateway disabled - set SSH_GATEWAY_ENABLE=true to enable")
	}

	// Create metrics collector
	metricsCollector := metrics.NewCollector(metrics.CollectorConfig{
		Logger:                             logger,
		Docker:                             dockerClient,
		WindowSize:                         cfg.CollectorWindowSize,
		CPUUsageSnapshotInterval:           cfg.CPUUsageSnapshotInterval,
		AllocatedResourcesSnapshotInterval: cfg.AllocatedResourcesSnapshotInterval,
	})
	metricsCollector.Start(ctx)

	_, err = runner.GetInstance(&runner.RunnerInstanceConfig{
		Logger:             logger,
		BackupInfoCache:    backupInfoCache,
		SnapshotErrorCache: cache.NewSnapshotErrorCache(ctx, cfg.SnapshotErrorCacheRetention),
		Docker:             dockerClient,
		SandboxService:     sandboxService,
		MetricsCollector:   metricsCollector,
		NetRulesManager:    netRulesManager,
		SSHGatewayService:  sshGatewayService,
	})
	if err != nil {
		logger.Error("Failed to initialize runner instance", "error", err)
		return 2
	}

	if cfg.ApiVersion == 2 {
		healthcheckService, err := healthcheck.NewService(&healthcheck.HealthcheckServiceConfig{
			Interval:   cfg.HealthcheckInterval,
			Timeout:    cfg.HealthcheckTimeout,
			Collector:  metricsCollector,
			Logger:     logger,
			Domain:     cfg.Domain,
			ApiPort:    cfg.ApiPort,
			ProxyPort:  cfg.ApiPort,
			TlsEnabled: cfg.EnableTLS,
			Docker:     dockerClient,
		})
		if err != nil {
			logger.Error("Failed to create healthcheck service", "error", err)
			return 2
		}

		go func() {
			logger.Info("Starting healthcheck service")
			healthcheckService.Start(ctx)
		}()

		executorService, err := executor.NewExecutor(&executor.ExecutorConfig{
			Logger:    logger,
			Docker:    dockerClient,
			Collector: metricsCollector,
		})
		if err != nil {
			logger.Error("Failed to create executor service", "error", err)
			return 2
		}

		pollerService, err := poller.NewService(&poller.PollerServiceConfig{
			PollTimeout: cfg.PollTimeout,
			PollLimit:   cfg.PollLimit,
			Logger:      logger,
			Executor:    executorService,
		})
		if err != nil {
			logger.Error("Failed to create poller service", "error", err)
			return 2
		}

		go func() {
			logger.Info("Starting poller service")
			pollerService.Start(ctx)
		}()
	}

	apiServer := api.NewApiServer(api.ApiServerConfig{
		Logger:      logger,
		ApiPort:     cfg.ApiPort,
		ApiToken:    cfg.ApiToken,
		TLSCertFile: cfg.TLSCertFile,
		TLSKeyFile:  cfg.TLSKeyFile,
		EnableTLS:   cfg.EnableTLS,
		LogRequests: cfg.ApiLogRequests,
	})

	apiServerErrChan := make(chan error)

	go func() {
		err := apiServer.Start(ctx)
		apiServerErrChan <- err
	}()

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-apiServerErrChan:
		logger.Error("API server error", "error", err)
		return 1
	case <-interruptChannel:
		logger.Info("Signal received, shutting down")
		apiServer.Stop()
		logger.Info("Shutdown complete")
		return 143 // SIGTERM
	case err := <-monitorErrChan:
		logger.Error("Docker monitor error", "error", err)
		return 1
	}
}

// maybeBuildInContainerMounter builds the layered in-container mounter when
// LAYERED_BINARY_PATH is set, or returns (nil, nil) to leave the backend
// disabled (layered sandboxes then fall back to s3fuse). It errors only when
// the path is set but invalid, so the runner fails fast on misconfiguration.
func maybeBuildInContainerMounter(_ context.Context, cfg *config.Config, logger *slog.Logger) (volume.Mounter, error) {
	if cfg.LayeredBinaryPath == "" {
		return nil, nil
	}
	if _, err := os.Stat(cfg.LayeredBinaryPath); err != nil {
		return nil, fmt.Errorf("LAYERED_BINARY_PATH not accessible: %w", err)
	}

	mounter := incontainer.NewMounter(incontainer.Config{
		LayeredBinaryHostPath: cfg.LayeredBinaryPath,
	})

	logger.Info(
		"Layered in-container volume backend enabled",
		"layeredBinary", cfg.LayeredBinaryPath,
	)
	return mounter, nil
}
