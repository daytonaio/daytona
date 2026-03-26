// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/daytonaio/runner/pkg/volume/multiplexer"
)

func main() {
	// Parse command line flags
	var (
		mountPath   = flag.String("mount-path", "/mnt/daytona-volumes", "FUSE mount path")
		grpcAddress = flag.String("grpc-address", "unix:///var/run/daytona-volume-multiplexer.sock", "gRPC server address")
		cacheDir    = flag.String("cache-dir", "/var/cache/daytona-volumes", "Cache directory")
		maxCacheGB  = flag.Int("max-cache-gb", 10, "Maximum cache size in GB")
		logLevel    = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	)
	flag.Parse()

	// Set up logging
	level := slog.LevelInfo
	switch *logLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	// Create daemon
	daemon := multiplexer.NewMultiplexerDaemon(*mountPath, *cacheDir, logger)
	_ = maxCacheGB // TODO: Use this to configure cache size

	// Start gRPC server
	grpcServer := multiplexer.NewGRPCServer(daemon, logger)
	go func() {
		if err := grpcServer.Start(*grpcAddress); err != nil {
			logger.Error("Failed to start gRPC server", "error", err)
			os.Exit(1)
		}
	}()

	// Handle shutdown signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Start FUSE daemon
	if err := daemon.Start(ctx); err != nil {
		logger.Error("Failed to start daemon", "error", err)
		os.Exit(1)
	}

	// Clean shutdown
	grpcServer.Stop()
	if err := daemon.Stop(); err != nil {
		logger.Error("Failed to stop daemon cleanly", "error", err)
	}

	logger.Info("Volume multiplexer shutdown complete")
}

// Factory function for creating providers
func init() {
	// Register provider factories here
	// This would normally be in the multiplexer package but shown here for clarity

	// Example:
	// multiplexer.RegisterProviderFactory("s3", func() volume.Provider {
	//     return s3.NewS3Provider()
	// })
}
