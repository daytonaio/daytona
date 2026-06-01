// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daytonaio/common-go/pkg/log"
	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/internal"
	"github.com/daytonaio/runner/internal/metrics"
	"github.com/daytonaio/runner/pkg/api"
	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/daemon"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/netrules"
	otelpkg "github.com/daytonaio/runner/pkg/otel"
	"github.com/daytonaio/runner/pkg/runner"
	"github.com/daytonaio/runner/pkg/runner/v2/executor"
	"github.com/daytonaio/runner/pkg/runner/v2/healthcheck"
	"github.com/daytonaio/runner/pkg/runner/v2/poller"
	"github.com/daytonaio/runner/pkg/services"
	"github.com/daytonaio/runner/pkg/sshgateway"
	"github.com/daytonaio/runner/pkg/telemetry/filters"
	"github.com/docker/docker/client"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	os.Exit(run())
}

func run() int {
	logLevel := log.ParseLogLevel(os.Getenv("LOG_LEVEL"))

	// Console-only logger so config errors (and anything before otel.Init) are visible.
	otelpkg.InitConsoleLogging(logLevel)
	logger := slog.Default()

	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error("Failed to get config", "error", err)
		return 2
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Translate the legacy per-signal toggles to the standard OTel env vars
	// consumed by the autoexport-based pkg/otel.Init. The runner historically
	// required OTEL_LOGGING_ENABLED / OTEL_TRACING_ENABLED to be explicitly true
	// even when an endpoint was configured; preserve that opt-in default by
	// disabling the respective signal exporters when the flag is unset/false.
	if !cfg.OtelLoggingEnabled && os.Getenv("OTEL_LOGS_EXPORTER") == "" {
		_ = os.Setenv("OTEL_LOGS_EXPORTER", "none")
	}
	if !cfg.OtelTracingEnabled && os.Getenv("OTEL_TRACES_EXPORTER") == "" {
		_ = os.Setenv("OTEL_TRACES_EXPORTER", "none")
	}
	if !cfg.OtelMetricsEnabled && os.Getenv("OTEL_METRICS_EXPORTER") == "" {
		_ = os.Setenv("OTEL_METRICS_EXPORTER", "none")
	}

	otelShutdown, err := otelpkg.Init(ctx, "daytona-runner", internal.Version, cfg.Environment, logLevel,
		otelpkg.WithSpanExporterWrapper(func(exp sdktrace.SpanExporter) sdktrace.SpanExporter {
			return (&filters.NotFoundExporterFilter{}).Apply(exp)
		}),
	)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to initialize OpenTelemetry", "error", err)
		return 2
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := otelShutdown(shutdownCtx); err != nil {
			slog.Warn("otel shutdown error", "error", err)
		}
	}()

	// pkg/otel may have swapped the default logger to the OTel-fanout handler.
	logger = slog.Default()

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

	dockerClient, err := docker.NewDockerClient(ctx, docker.DockerClientConfig{
		ApiClient:                    cli,
		BackupInfoCache:              backupInfoCache,
		Logger:                       logger,
		AWSRegion:                    cfg.AWSRegion,
		AWSEndpointUrl:               cfg.AWSEndpointUrl,
		AWSAccessKeyId:               cfg.AWSAccessKeyId,
		AWSSecretAccessKey:           cfg.AWSSecretAccessKey,
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
