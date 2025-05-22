// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/daytonaio/runner-docker/cmd/config"
	"github.com/daytonaio/runner-docker/internal/util"
	"github.com/daytonaio/runner-docker/pkg/cache"
	"github.com/daytonaio/runner-docker/pkg/daemon"
	"github.com/daytonaio/runner-docker/pkg/models"
	"github.com/daytonaio/runner-docker/pkg/server"
	"github.com/daytonaio/runner-docker/pkg/server/middlewares"
	pb "github.com/daytonaio/runner/proto"
	"github.com/docker/docker/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"

	log "github.com/sirupsen/logrus"

	golog "log"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("Config loaded")

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("Docker APIClient created")

	cache := cache.NewInMemoryRunnerCache(cache.InMemoryRunnerCacheConfig{
		Cache:         make(map[string]*models.CacheData),
		RetentionDays: cfg.CacheRetentionDays,
	})

	log.Info("Cache created")

	// Start cleanup job with a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache.Cleanup(ctx)

	log.Info("Cache cleanup started")

	daemonPath, err := daemon.WriteDaemonBinary()
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("Daemon copied")

	runner := server.NewRunnerServer(server.RunnerServerConfig{
		ApiClient:          cli,
		Cache:              cache,
		LogWriter:          os.Stdout,
		AWSRegion:          cfg.AWSRegion,
		AWSEndpointUrl:     cfg.AWSEndpointUrl,
		AWSAccessKeyId:     cfg.AWSAccessKeyId,
		AWSSecretAccessKey: cfg.AWSSecretAccessKey,
		DaemonPath:         daemonPath,
	})

	log.Info("Created runner")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			middlewares.ChainUnaryServer(
				middlewares.GetDefaultInterceptors()...,
			),
		),
	)
	if cfg.EnableTLS {
		// Load TLS certificates
		cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			log.Fatalf("failed to load certificates: %v", err)
		}

		// Create TLS config
		tlsCfg := &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.NoClientCert, // or tls.RequireAndVerifyClientCert for mutual TLS
		}

		// Create credentials
		creds := credentials.NewTLS(tlsCfg)

		// Create gRPC server with TLS
		s = grpc.NewServer(
			grpc.Creds(creds),
			grpc.UnaryInterceptor(
				middlewares.ChainUnaryServer(
					middlewares.GetDefaultInterceptors()...,
				),
			),
		)
	}

	log.Info("Created gRPC server with runner")

	pb.RegisterRunnerServer(s, runner)

	// // Start Prometheus metrics endpoint
	// go func() {
	//     http.Handle("/metrics", promhttp.Handler())
	//     log.Fatal(http.ListenAndServe(":9090", nil))
	// }()

	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func init() {
	logLevel := log.WarnLevel

	logLevelEnv, logLevelSet := os.LookupEnv("LOG_LEVEL")

	if logLevelSet {
		var err error
		logLevel, err = log.ParseLevel(logLevelEnv)
		if err != nil {
			logLevel = log.WarnLevel
		}
	}

	log.SetLevel(logLevel)

	log.SetOutput(os.Stdout)

	logFilePath, logFilePathSet := os.LookupEnv("LOG_FILE_PATH")
	if logFilePathSet {
		logDir := filepath.Dir(logFilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Error("Failed to create log directory:", err)
			os.Exit(1)
		}

		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		log.SetOutput(io.MultiWriter(os.Stdout, file))
	}

	zerologLevel, err := zerolog.ParseLevel(logLevel.String())
	if err != nil {
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
