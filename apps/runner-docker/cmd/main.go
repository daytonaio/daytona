// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daytonaio/runner-docker/internal/config"
	"github.com/daytonaio/runner-docker/pkg/cache"
	"github.com/daytonaio/runner-docker/pkg/daemon"
	"github.com/daytonaio/runner-docker/pkg/grpc"
	"github.com/daytonaio/runner-docker/pkg/proxy"
	"github.com/docker/docker/client"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	// Setup logging
	log := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
		TimeFormat: time.RFC3339,
	}))

	// Load config
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error("Error loading config", "error", err)
		return
	}

	// Create Docker APIClient
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error("Error creating Docker APIClient", "error", err)
		return
	}

	// Create cache
	cache := cache.NewInMemoryRunnerCache(cache.InMemoryRunnerCacheConfig{
		Cache:         make(map[string]*cache.CacheData),
		RetentionDays: cfg.CacheRetentionDays,
	})

	cache.Cleanup(context.Background())

	daemonPath, err := daemon.WriteDaemonBinary()
	if err != nil {
		log.Error("Error writing daemon binary", "error", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create gRPC server
	grpcServer := grpc.New(grpc.ServerConfig{
		Address:            fmt.Sprintf(":%d", cfg.Port),
		DockerClient:       dockerClient,
		RunnerCache:        &cache,
		DaemonPath:         daemonPath,
		AWSAccessKeyId:     cfg.AWSAccessKeyId,
		AWSSecretAccessKey: cfg.AWSSecretAccessKey,
		AWSRegion:          cfg.AWSRegion,
		AWSEndpointUrl:     cfg.AWSEndpointUrl,
		Log:                log,
	})

	// Create metrics server
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.MetricsPort),
		Handler: promhttp.Handler(),
	}

	// Create proxy server
	proxyServer := proxy.New(proxy.ProxyServerConfig{
		DockerClient: dockerClient,
		Port:         cfg.ProxyPort,
		Log:          log,
	})

	errChan := make(chan error, 3)

	// Start gRPC server
	go func() {
		log.Info("Starting gRPC server", "address", fmt.Sprintf(":%d", cfg.Port))
		errChan <- grpcServer.Start()
	}()

	// Start proxy server
	go func() {
		log.Info("Starting proxy server", "address", fmt.Sprintf(":%d", cfg.ProxyPort))
		errChan <- proxyServer.Start()
	}()

	// Start metrics server
	go func() {
		log.Info("Starting metrics server", "address", fmt.Sprintf(":%d", cfg.MetricsPort))
		errChan <- metricsServer.ListenAndServe()
	}()

	// Blocking main and waiting for shutdown.
	select {
	case <-ctx.Done():
		log.Info("Received shutdown signal.")
	case err := <-errChan:
		if err != nil {
			log.Error("Server error", "error", err)
		} else {
			log.Info("Server exited cleanly.")
		}
		stop() // stop signal.NotifyContext
	}

	// Graceful shutdown with timeout enforcement
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a channel to track shutdown completion
	shutdownDone := make(chan struct{})

	// Start shutdown in a goroutine
	go func() {
		log.Info("Shutting down gRPC server...")
		grpcServer.Shutdown(shutdownCtx)

		log.Info("Shutting down Proxy server...")
		if err := proxyServer.Shutdown(shutdownCtx); err != nil {
			log.Error("Proxy server shutdown error", "error", err)
		}

		log.Info("Shutting down Metrics server...")
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			log.Error("Metrics server shutdown error", "error", err)
		}

		close(shutdownDone)
	}()

	// Wait for either shutdown completion or timeout
	select {
	case <-shutdownDone:
		log.Info("All servers shut down gracefully")
	case <-shutdownCtx.Done():
		log.Error("Shutdown timeout reached, forcing exit")
		os.Exit(1)
	}
}
