// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	golog "log"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/internal/util"
	"github.com/daytonaio/runner/pkg/api"
	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/daemon"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/models"
	"github.com/daytonaio/runner/pkg/netrules"
	"github.com/daytonaio/runner/pkg/runner"
	"github.com/daytonaio/runner/pkg/services"
	"github.com/daytonaio/runner/pkg/sshgateway"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"

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

	apiServer := api.NewApiServer(api.ApiServerConfig{
		ApiPort:     cfg.ApiPort,
		TLSCertFile: cfg.TLSCertFile,
		TLSKeyFile:  cfg.TLSKeyFile,
		EnableTLS:   cfg.EnableTLS,
	})

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Errorf("Error creating Docker client: %v", err)
		return
	}

	// Initialize net rules manager
	persistent := cfg.Environment == "production"
	netRulesManager, err := netrules.NewNetRulesManager(persistent)
	if err != nil {
		log.Error(err)
		return
	}

	// Start Docker events monitor
	monitor := docker.NewDockerMonitor(cli, netRulesManager)
	go func() {
		err = monitor.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()
	defer monitor.Stop()

	runnerCache := cache.NewInMemoryRunnerCache(cache.InMemoryRunnerCacheConfig{
		Cache:         make(map[string]*models.CacheData),
		RetentionDays: cfg.CacheRetentionDays,
	})

	// Start cleanup job with a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runnerCache.Cleanup(ctx)

	daemonPath, err := daemon.WriteStaticBinary("daemon-amd64")
	if err != nil {
		log.Errorf("Error writing daemon binary: %v", err)
		return
	}

	pluginPath, err := daemon.WriteStaticBinary("daytona-computer-use")
	if err != nil {
		log.Errorf("Error writing plugin binary: %v", err)
		return
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient:             cli,
		Cache:                 runnerCache,
		LogWriter:             os.Stdout,
		AWSRegion:             cfg.AWSRegion,
		AWSEndpointUrl:        cfg.AWSEndpointUrl,
		AWSAccessKeyId:        cfg.AWSAccessKeyId,
		AWSSecretAccessKey:    cfg.AWSSecretAccessKey,
		DaemonPath:            daemonPath,
		ComputerUsePluginPath: pluginPath,
		NetRulesManager:       netRulesManager,
	})

	sandboxService := services.NewSandboxService(runnerCache, dockerClient)

	metricsCache := cache.NewMapCache[models.SystemMetrics]()

	metricsService := services.NewMetricsService(services.MetricsServiceConfig{
		Docker:   dockerClient,
		Cache:    metricsCache,
		Interval: 15 * time.Second,
	})
	metricsService.StartMetricsCollection(ctx)

	// Initialize SSH Gateway if enabled
	var sshGatewayService *sshgateway.Service
	if sshgateway.IsSSHGatewayEnabled() {
		sshGatewayService = sshgateway.NewService(dockerClient)

		go func() {
			log.Info("Starting SSH Gateway")
			if err := sshGatewayService.Start(ctx); err != nil {
				log.Errorf("SSH Gateway error: %v", err)
			}
		}()
	} else {
		log.Info("Gateway disabled - set SSH_GATEWAY_ENABLE=true to enable")
	}

	_ = runner.GetInstance(&runner.RunnerInstanceConfig{
		Cache:             runnerCache,
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
		log.Errorf("API server error: %v", err)
		return
	case <-interruptChannel:
		apiServer.Stop()
	}
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
