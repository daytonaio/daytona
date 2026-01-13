/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/daytonaio/snapshot-manager/internal/config"
	"github.com/daytonaio/snapshot-manager/internal/logger"
	"github.com/daytonaio/snapshot-manager/internal/server"
)

func main() {
	log := logger.NewLogger()

	cfg, err := config.Load()
	if err != nil {
		log.Error("Failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	serverCfg, err := server.BuildConfig(cfg)
	if err != nil {
		log.Error("Failed to build server configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	serverInstance, err := server.NewServer(serverCfg, log)
	if err != nil {
		log.Error("Failed to create server instance", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info("Received shutdown signal", slog.String("signal", sig.String()))
		cancel()
	}()

	if err := serverInstance.Start(ctx); err != nil {
		log.Error("Registry error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("Daytona Snapshot manager stopped")
}
