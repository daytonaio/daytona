// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/internal/util"
	"github.com/daytonaio/runner/pkg/api"
	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/daemon"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/netrules"
	"github.com/daytonaio/runner/pkg/runner"
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

	apiServer := api.NewApiServer(api.ApiServerConfig{
		ApiPort:     cfg.ApiPort,
		TLSCertFile: cfg.TLSCertFile,
		TLSKeyFile:  cfg.TLSKeyFile,
		EnableTLS:   cfg.EnableTLS,
	})

	// Init tracing
	shutdownTracing, err := telemetry.InitTracing(cfg)
	if err != nil {
		logger.Error("Failed to initialize tracing", "error", err)
		return
	}
	defer shutdownTracing()

	// Init logging
	shutdownLogging, err := telemetry.InitLogging(logger, cfg)
	if err != nil {
		logger.Error("Failed to initialize OTEL logging", "error", err)
		return
	}
	defer shutdownLogging()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation(), client.WithTraceProvider(otel.GetTracerProvider()))
	if err != nil {
		logger.Error("Error creating Docker client", "error", err)
		return
	}

	// Initialize net rules manager
	persistent := cfg.Environment != "development"
	netRulesManager, err := netrules.NewNetRulesManager(persistent)
	if err != nil {
		logger.Error("Failed to initialize net rules manager", "error", err)
		return
	}

	// Start net rules manager
	if err = netRulesManager.Start(); err != nil {
		logger.Error("Failed to start net rules manager", "error", err)
		return
	}

	// Start Docker events monitor
	monitor := docker.NewDockerMonitor(cli, netRulesManager)
	go func() {
		err = monitor.Start()
		if err != nil {
			logger.Error("Failed to start Docker events monitor", "error", err)
		}
	}()
	defer monitor.Stop()
	defer netRulesManager.Stop()

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
	defer cancel()

	statesCache := cache.GetStatesCache(cfg.CacheRetentionDays)

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient:              cli,
		StatesCache:            statesCache,
		AWSRegion:              cfg.AWSRegion,
		AWSEndpointUrl:         cfg.AWSEndpointUrl,
		AWSAccessKeyId:         cfg.AWSAccessKeyId,
		AWSSecretAccessKey:     cfg.AWSSecretAccessKey,
		DaemonPath:             daemonPath,
		ComputerUsePluginPath:  pluginPath,
		NetRulesManager:        netRulesManager,
		ResourceLimitsDisabled: cfg.ResourceLimitsDisabled,
	})

	sandboxService := services.NewSandboxService(statesCache, dockerClient)

	metricsService := services.NewMetricsService(services.MetricsServiceConfig{
		Docker:   dockerClient,
		Interval: 15 * time.Second,
	})
	metricsService.StartMetricsCollection(ctx)

	// Initialize sandbox state synchronization service
	sandboxSyncService := services.NewSandboxSyncService(services.SandboxSyncServiceConfig{
		Docker:   dockerClient,
		Interval: 10 * time.Second, // Sync every 10 seconds
	})
	sandboxSyncService.StartSyncProcess(ctx)

	// Initialize SSH Gateway if enabled
	var sshGatewayService *sshgateway.Service
	if sshgateway.IsSSHGatewayEnabled() {
		sshGatewayService = sshgateway.NewService(dockerClient)

		go func() {
			logger.Info("Starting SSH Gateway")
			if err := sshGatewayService.Start(ctx); err != nil {
				logger.Error("SSH Gateway error", "error", err)
			}
		}()
	} else {
		logger.Info("Gateway disabled - set SSH_GATEWAY_ENABLE=true to enable")
	}

	_ = runner.GetInstance(&runner.RunnerInstanceConfig{
		StatesCache:       statesCache,
		Docker:            dockerClient,
		SandboxService:    sandboxService,
		MetricsService:    metricsService,
		NetRulesManager:   netRulesManager,
		SSHGatewayService: sshGatewayService,
	})

	apiServerErrChan := make(chan error)

	go func() {
		err := apiServer.Start()
		apiServerErrChan <- err
	}()

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)

	select {
	case err := <-apiServerErrChan:
		logger.Error("API server error", "error", err)
		return
	case <-interruptChannel:
		apiServer.Stop()
	}
}
