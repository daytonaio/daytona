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

	"github.com/daytonaio/runner-ch/cmd/runner/config"
	"github.com/daytonaio/runner-ch/internal/metrics"
	"github.com/daytonaio/runner-ch/internal/util"
	"github.com/daytonaio/runner-ch/pkg/api"
	"github.com/daytonaio/runner-ch/pkg/cloudhypervisor"
	"github.com/daytonaio/runner-ch/pkg/netrules"
	"github.com/daytonaio/runner-ch/pkg/runner"
	"github.com/daytonaio/runner-ch/pkg/runner/v2/executor"
	"github.com/daytonaio/runner-ch/pkg/runner/v2/healthcheck"
	"github.com/daytonaio/runner-ch/pkg/runner/v2/poller"
	"github.com/daytonaio/runner-ch/pkg/sshgateway"
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

	// Initialize Cloud Hypervisor client
	chClient, err := cloudhypervisor.NewClient(cloudhypervisor.ClientConfig{
		SandboxesPath:   cfg.CHSandboxesPath,
		SnapshotsPath:   cfg.CHSnapshotsPath,
		SocketsPath:     cfg.CHSocketsPath,
		KernelPath:      cfg.CHKernelPath,
		InitramfsPath:   cfg.CHInitramfsPath,
		FirmwarePath:    cfg.CHFirmwarePath,
		BaseImagePath:   cfg.CHBaseImagePath,
		DefaultCpus:     cfg.CHDefaultCpus,
		DefaultMemoryMB: uint64(cfg.CHDefaultMemoryMB),
		SSHHost:         cfg.CHSSHHost,
		SSHKeyPath:      cfg.CHSSHKeyPath,
		BridgeName:      cfg.CHBridgeName,
		TapCreateScript: cfg.CHTapCreateScript,
		TapDeleteScript: cfg.CHTapDeleteScript,
		TapPoolEnabled:  cfg.TapPoolEnabled,
		TapPoolSize:     cfg.TapPoolSize,
	})
	if err != nil {
		log.Fatalf("Failed to create Cloud Hypervisor client: %v", err)
	}
	defer chClient.Close()

	// Start TAP pool if enabled
	if cfg.TapPoolEnabled {
		if err := chClient.StartTapPool(ctx); err != nil {
			log.Warnf("Failed to start TAP pool: %v", err)
		} else {
			log.Infof("TAP pool started (size=%d)", cfg.TapPoolSize)
		}
		defer chClient.StopTapPool()
	}

	// Initialize network namespace pool - loads existing allocations to prevent IP conflicts
	if err := chClient.InitializeNetNSPool(ctx); err != nil {
		log.Warnf("Failed to initialize NetNS pool: %v", err)
	}

	log.Info("Cloud Hypervisor runner started")
	log.Infof("Configuration:")
	log.Infof("  Sandboxes path: %s", cfg.CHSandboxesPath)
	log.Infof("  Snapshots path: %s", cfg.CHSnapshotsPath)
	log.Infof("  Sockets path: %s", cfg.CHSocketsPath)
	log.Infof("  Kernel path: %s", cfg.CHKernelPath)
	log.Infof("  Initramfs path: %s", cfg.CHInitramfsPath)
	log.Infof("  Base image: %s", cfg.CHBaseImagePath)
	log.Infof("  Bridge: %s", cfg.CHBridgeName)
	if cfg.CHSSHHost != "" {
		log.Infof("  Remote mode: SSH to %s", cfg.CHSSHHost)
	} else {
		log.Infof("  Local mode")
	}

	// Initialize IP pool (loads existing allocations)
	log.Info("Initializing IP pool...")
	if err := chClient.InitializeIPPool(ctx); err != nil {
		log.Warnf("Failed to initialize IP pool: %v", err)
	}

	// List existing sandboxes
	sandboxes, err := chClient.List(ctx)
	if err != nil {
		log.Warnf("Failed to list existing sandboxes: %v", err)
	} else {
		log.Infof("Found %d existing sandboxes", len(sandboxes))
	}

	// Initialize network rules manager
	netRulesManager, err := netrules.NewNetRulesManager("")
	if err != nil {
		log.Warnf("Failed to create network rules manager: %v", err)
		// Continue without network rules - not critical for basic operation
	}

	// Create runner instance
	r := runner.NewRunner(chClient, netRulesManager)

	// Setup structured logger
	slogLogger := newSLogger()

	// Determine if we're in local mode (no SSH host)
	localMode := cfg.CHSSHHost == ""

	// Initialize metrics collector
	metricsCollector := metrics.NewCollector(slogLogger, localMode, chClient)

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
			CHClient:  chClient,
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
	} else {
		log.Warn("No Daytona API URL configured - running in standalone mode (no job polling or healthcheck)")
	}

	// Start SSH gateway if enabled
	if sshgateway.IsSSHGatewayEnabled() {
		sshGatewayService := sshgateway.NewService(chClient)
		go func() {
			// Note: Start() logs "SSH Gateway listening on port X" on success
			if err := sshGatewayService.Start(ctx); err != nil {
				log.Errorf("SSH Gateway error: %v", err)
			}
		}()
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
