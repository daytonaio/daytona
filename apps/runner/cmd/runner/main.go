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

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/internal/metrics"
	"github.com/daytonaio/runner/internal/util"
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
	"github.com/daytonaio/runner/pkg/telemetry"
	"github.com/docker/docker/client"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"go.opentelemetry.io/otel"
)

func main() {
	// Init slog logger
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
		TimeFormat: time.RFC3339,
		Level:      util.ParseLogLevel(os.Getenv("LOG_LEVEL")),
	}))

	slog.SetDefault(logger)

	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error("Failed to get config", "error", err)
		return
	}

	// Init tracing
	shutdownTracing, err := telemetry.InitTracing(telemetry.OtelTracingConfig{
		OtelTracingEnabled:  cfg.OtelTracingEnabled,
		OtelSampleRate:      cfg.OtelSampleRate,
		OtelBatchTimeout:    cfg.OtelBatchTimeout,
		OtelMaxBatchSize:    cfg.OtelMaxBatchSize,
		OtlpExporterTimeout: cfg.OtlpExporterTimeout,
		Environment:         cfg.Environment,
	})
	if err != nil {
		logger.Error("Failed to initialize tracing", "error", err)
		return
	}

	// Init logging
	shutdownLogging, err := telemetry.InitLogging(telemetry.OtelLoggingConfig{
		Logger:              logger,
		OtelLoggingEnabled:  cfg.OtelLoggingEnabled,
		OtlpExporterTimeout: cfg.OtlpExporterTimeout,
		Environment:         cfg.Environment,
	})
	if err != nil {
		logger.Error("Failed to initialize OTEL logging", "error", err)
		return
	}

	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
		client.WithTraceProvider(otel.GetTracerProvider()),
	)
	if err != nil {
		logger.Error("Error creating Docker client", "error", err)
		return
	}

	// Initialize net rules manager
	persistent := cfg.Environment != "development"
	netRulesManager, err := netrules.NewNetRulesManager(logger, persistent)
	if err != nil {
		logger.Error("Failed to initialize net rules manager", "error", err)
		return
	}

	// Start net rules manager
	if err = netRulesManager.Start(); err != nil {
		logger.Error("Failed to start net rules manager", "error", err)
		return
	}

	daemonPath, err := daemon.WriteStaticBinary("daemon-amd64")
	if err != nil {
		logger.Error("Error writing daemon binary", "error", err)
		return
	}

	pluginPath, err := daemon.WriteStaticBinary("daytona-computer-use")
	if err != nil {
		logger.Error("Error writing plugin binary", "error", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	statesCache := cache.GetStatesCache(cfg.CacheRetentionDays)

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient:                cli,
		Logger:                   logger,
		StatesCache:              statesCache,
		AWSRegion:                cfg.AWSRegion,
		AWSEndpointUrl:           cfg.AWSEndpointUrl,
		AWSAccessKeyId:           cfg.AWSAccessKeyId,
		AWSSecretAccessKey:       cfg.AWSSecretAccessKey,
		DaemonPath:               daemonPath,
		ComputerUsePluginPath:    pluginPath,
		NetRulesManager:          netRulesManager,
		ResourceLimitsDisabled:   cfg.ResourceLimitsDisabled,
		UseSnapshotEntrypoint:    cfg.UseSnapshotEntrypoint,
		VolumeCleanupIntervalSec: cfg.VolumeCleanupIntervalSec,
		BackupTimeoutMin:         cfg.BackupTimeoutMin,
	})

	// Start Docker events monitor
	monitorOpts := docker.MonitorOptions{
		OnDestroyEvent: func(ctx context.Context) {
			dockerClient.CleanupOrphanedVolumeMounts(ctx)
		},
	}
	monitor := docker.NewDockerMonitor(logger, cli, netRulesManager, monitorOpts)
	go func() {
		err = monitor.Start()
		if err != nil {
			logger.Error("Failed to start Docker events monitor", "error", err)
		}
	}()

	sandboxService := services.NewSandboxService(logger, statesCache, dockerClient)

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
	metricsCollector := metrics.NewCollector(logger, dockerClient, cfg.CollectorWindowSize)
	metricsCollector.Start(ctx)

	_ = runner.GetInstance(&runner.RunnerInstanceConfig{
		StatesCache:       statesCache,
		Docker:            dockerClient,
		SandboxService:    sandboxService,
		MetricsCollector:  metricsCollector,
		NetRulesManager:   netRulesManager,
		SSHGatewayService: sshGatewayService,
	})

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
		})
		if err != nil {
			logger.Error("Failed to create healthcheck service", "error", err)
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
		}

		pollerService, err := poller.NewService(&poller.PollerServiceConfig{
			PollTimeout: cfg.PollTimeout,
			PollLimit:   cfg.PollLimit,
			Logger:      logger,
			Executor:    executorService,
		})
		if err != nil {
			logger.Error("Failed to create poller service", "error", err)
		}

		go func() {
			logger.Info("Starting poller service")
			pollerService.Start(ctx)
			if err != nil {
				logger.Error("Poller service error", "error", err)
			}
		}()
	}

	apiServer := api.NewApiServer(api.ApiServerConfig{
		ApiPort:     cfg.ApiPort,
		ApiToken:    cfg.ApiToken,
		TLSCertFile: cfg.TLSCertFile,
		TLSKeyFile:  cfg.TLSKeyFile,
		EnableTLS:   cfg.EnableTLS,
	})

	apiServerErrChan := make(chan error)

	go func() {
		err := apiServer.Start()
		apiServerErrChan <- err
	}()

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-apiServerErrChan:
		logger.Error("API server error", "error", err)
		return
	case <-interruptChannel:
		logger.Info("Signal received, shutting down")

		cancel()

		monitor.Stop()
		netRulesManager.Stop()
		apiServer.Stop()

		shutdownLogging()
		shutdownTracing()

		logger.Info("Shutdown complete")
	}
}
