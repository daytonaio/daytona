// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	golog "log"

	"github.com/daytonaio/runner-android/cmd/runner/config"
	"github.com/daytonaio/runner-android/internal/metrics"
	"github.com/daytonaio/runner-android/internal/util"
	"github.com/daytonaio/runner-android/pkg/api"
	"github.com/daytonaio/runner-android/pkg/cuttlefish"
	"github.com/daytonaio/runner-android/pkg/runner"
	"github.com/daytonaio/runner-android/pkg/runner/v2/executor"
	"github.com/daytonaio/runner-android/pkg/runner/v2/healthcheck"
	"github.com/daytonaio/runner-android/pkg/runner/v2/poller"
	"github.com/daytonaio/runner-android/pkg/sshgateway"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"

	log "github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Errorf("Failed to get config: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Cuttlefish client
	cvdClient, err := cuttlefish.NewClient(cuttlefish.ClientConfig{
		InstancesPath:   cfg.CVDInstancesPath,
		ArtifactsPath:   cfg.CVDArtifactsPath,
		CVDHome:         cfg.CVDHome,
		LaunchCVDPath:   cfg.CVDLaunchPath,
		StopCVDPath:     cfg.CVDStopPath,
		CVDPath:         cfg.CVDPath,
		ADBPath:         cfg.CVDADBPath,
		DefaultCpus:     cfg.CVDDefaultCpus,
		DefaultMemoryMB: uint64(cfg.CVDDefaultMemoryMB),
		DefaultDiskGB:   cfg.CVDDefaultDiskGB,
		BaseInstanceNum: cfg.CVDBaseInstance,
		MaxInstances:    cfg.CVDMaxInstances,
		SSHHost:         cfg.CVDSSHHost,
		SSHKeyPath:      cfg.CVDSSHKeyPath,
		ADBBasePort:     cfg.CVDADBBasePort,
		WebRTCBasePort:  cfg.CVDWebRTCBasePort,
	})
	if err != nil {
		log.Fatalf("Failed to create Cuttlefish client: %v", err)
	}
	defer cvdClient.Close()

	log.Info("Cuttlefish Android runner started")
	log.Infof("Configuration:")
	log.Infof("  Instances path: %s", cfg.CVDInstancesPath)
	log.Infof("  Artifacts path: %s", cfg.CVDArtifactsPath)
	log.Infof("  CVD home: %s", cfg.CVDHome)
	log.Infof("  ADB base port: %d", cfg.CVDADBBasePort)
	log.Infof("  Max instances: %d", cfg.CVDMaxInstances)
	if cfg.CVDSSHHost != "" {
		log.Infof("  Remote mode: SSH to %s", cfg.CVDSSHHost)
	} else {
		log.Infof("  Local mode")
	}

	// Recover orphaned sandboxes (runner-side)
	log.Info("Checking for orphaned sandboxes to recover...")
	if err := cvdClient.RecoverOrphanedSandboxes(ctx); err != nil {
		log.Warnf("Failed to recover orphaned sandboxes: %v", err)
	}

	// Sync CVD state - remove stale CVD instances not tracked by runner
	log.Info("Synchronizing CVD state with runner state...")
	if err := cvdClient.SyncCVDState(ctx); err != nil {
		log.Warnf("Failed to sync CVD state: %v", err)
	}

	// List existing sandboxes
	sandboxes, err := cvdClient.List(ctx)
	if err != nil {
		log.Warnf("Failed to list existing sandboxes: %v", err)
	} else {
		log.Infof("Found %d existing sandboxes", len(sandboxes))
	}

	// Create runner instance
	r := runner.NewRunner(cvdClient, nil)

	// Setup structured logger
	slogLogger := newSLogger()

	// Determine if we're in local mode (no SSH host)
	localMode := cfg.CVDSSHHost == ""

	// Initialize metrics collector
	metricsCollector := metrics.NewCollector(slogLogger, localMode, cvdClient)

	// Start API server in a goroutine
	apiServer := api.NewApiServer(api.ApiServerConfig{
		ApiPort:     cfg.ApiPort,
		TLSCertFile: cfg.TLSCertFile,
		TLSKeyFile:  cfg.TLSKeyFile,
		EnableTLS:   cfg.EnableTLS,
	}, r)

	go func() {
		if err := apiServer.Start(); err != nil {
			log.Fatalf("Failed to start API server: %v", err)
		}
	}()

	log.Infof("API server listening on port %d", cfg.ApiPort)

	// Initialize v2 services if API URL is configured
	if cfg.ServerUrl != "" {
		log.Info("Daytona API URL configured, starting v2 services...")

		// Create executor
		exec, err := executor.NewExecutor(&executor.ExecutorConfig{
			CVDClient: cvdClient,
			Collector: metricsCollector,
			Logger:    slogLogger,
		})
		if err != nil {
			log.Fatalf("Failed to create executor: %v", err)
		}

		// Create and start poller
		pollerService, err := poller.NewService(&poller.PollerServiceConfig{
			PollTimeout: cfg.PollTimeout,
			PollLimit:   cfg.PollLimit,
			Logger:      slogLogger,
			Executor:    exec,
		})
		if err != nil {
			log.Fatalf("Failed to create poller service: %v", err)
		}

		// Start poller in background
		go pollerService.Start(ctx)
		log.Info("Job poller started")

		// Create and start healthcheck
		healthcheckService, err := healthcheck.NewService(&healthcheck.HealthcheckServiceConfig{
			Interval:        cfg.HealthcheckInterval,
			Timeout:         cfg.HealthcheckTimeout,
			Collector:       metricsCollector,
			Logger:          slogLogger,
			Domain:          cfg.Domain,
			ProxyPort:       cfg.ApiPort,
			ProxyTLSEnabled: cfg.EnableTLS,
		})
		if err != nil {
			log.Fatalf("Failed to create healthcheck service: %v", err)
		}

		// Start healthcheck in background
		go healthcheckService.Start(ctx)
		log.Info("Healthcheck service started")

		// Create and start CVD health monitor to detect crashed instances
		cvdHealthMonitor, err := cuttlefish.NewHealthMonitor(cvdClient, &cuttlefish.HealthMonitorConfig{
			Interval:   30 * time.Second,
			MaxRetries: 2, // Report crash after 2 consecutive failed checks (~1 minute)
		})
		if err != nil {
			log.Warnf("Failed to create CVD health monitor: %v", err)
		} else {
			// Connect health monitor to CVD client for notifications
			cvdClient.SetHealthMonitor(cvdHealthMonitor)
			cvdHealthMonitor.Start(ctx)
			log.Info("CVD health monitor started")
		}
	} else {
		log.Warn("No Daytona API URL configured - running in standalone mode (no job polling or healthcheck)")
	}

	// Start SSH gateway for ADB tunneling (if enabled)
	if sshgateway.IsSSHGatewayEnabled() {
		sshGatewayService := sshgateway.NewService(cvdClient)
		go func() {
			if err := sshGatewayService.Start(ctx); err != nil {
				log.Errorf("SSH Gateway error: %v", err)
			}
		}()
		log.Infof("SSH Gateway started on port %d", sshGatewayService.GetPort())
	} else {
		log.Info("SSH Gateway disabled (set SSH_GATEWAY_ENABLE=true to enable)")
	}

	log.Info("Runner is ready!")

	// Wait for interrupt signal
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)

	<-interruptChannel
	log.Info("Shutting down...")

	// Cancel context to stop background services
	cancel()

	// Stop API server
	apiServer.Stop()
}

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
		// Continue anyway, as environment variables might be set directly
	}

	logLevel := log.WarnLevel

	logLevelEnv, logLevelSet := os.LookupEnv("LOG_LEVEL")

	if logLevelSet {
		var err error
		logLevel, err = log.ParseLevel(logLevelEnv)
		if err != nil {
			log.Warnf("Failed to parse log level '%s', using WarnLevel: %v", logLevelEnv, err)
			logLevel = log.WarnLevel
		}
	}

	log.SetLevel(logLevel)
	log.SetOutput(os.Stdout)

	logFilePath, logFilePathSet := os.LookupEnv("LOG_FILE_PATH")
	if logFilePathSet {
		logDir := filepath.Dir(logFilePath)

		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Errorf("Failed to create log directory: %v", err)
			os.Exit(1)
		}

		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Errorf("Failed to open log file: %v", err)
			os.Exit(1)
		}

		log.SetOutput(io.MultiWriter(os.Stdout, file))
	}

	zerologLevel, err := zerolog.ParseLevel(logLevel.String())
	if err != nil {
		log.Warnf("Failed to parse zerolog level, using ErrorLevel: %v", err)
		zerologLevel = zerolog.ErrorLevel
	}

	zerolog.SetGlobalLevel(zerologLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{
		Out:        &util.DebugLogWriter{},
		TimeFormat: time.RFC3339,
	})

	golog.SetOutput(&util.DebugLogWriter{})
}

func newSLogger() *slog.Logger {
	log := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
		TimeFormat: time.RFC3339,
		Level:      parseLogLevel(os.Getenv("LOG_LEVEL")),
	}))
	slog.SetDefault(log)
	return log
}

// parseLogLevel converts a string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
