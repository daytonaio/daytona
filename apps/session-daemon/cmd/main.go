// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/daytonaio/session-daemon/internal/config"
	"github.com/daytonaio/session-daemon/internal/logger"
	"github.com/daytonaio/session-daemon/internal/server"
)

func main() {
	log := logger.NewLogger()

	cfg, err := config.Load()
	if err != nil {
		log.Error("Failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	srv, err := server.NewServer(cfg, log)
	if err != nil {
		log.Error("Failed to create server", slog.String("error", err.Error()))
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

	if err := srv.Run(ctx); err != nil {
		log.Error("session-daemon error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("Daytona session-daemon stopped")
}
