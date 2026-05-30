//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	golog "log"

	"github.com/daytonaio/common-go/pkg/log"
	"github.com/daytonaio/daemon/cmd/daemon/config"
	"github.com/daytonaio/daemon/pkg/toolbox"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func main() {
	os.Exit(run())
}

func run() int {
	logLevel := log.ParseLogLevel(os.Getenv("LOG_LEVEL"))

	consoleHandler := tint.NewHandler(os.Stdout, &tint.Options{
		NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
		TimeFormat: time.RFC3339,
		Level:      logLevel,
	})

	logger := slog.New(consoleHandler)
	slog.SetDefault(logger)
	golog.SetOutput(&log.DebugLogWriter{})

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed to get user home directory", "error", err)
		return 2
	}

	configDir := filepath.Join(homeDir, ".daytona")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		logger.Error("Failed to create config directory", "path", configDir, "error", err)
		return 2
	}

	c, err := config.GetConfig()
	if err != nil {
		logger.Error("Failed to get config", "error", err)
		return 2
	}

	workDir, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get current working directory", "error", err)
		return 2
	}

	toolBoxServer := toolbox.NewServer(toolbox.ServerConfig{
		Logger:         logger,
		WorkDir:        workDir,
		ConfigDir:      configDir,
		OtelEndpoint:   c.OtelEndpoint,
		SandboxId:      c.SandboxId,
		OrganizationId: c.OrganizationId,
		RegionId:       c.RegionId,
		Snapshot:       c.Snapshot,
	})

	errChan := make(chan error)
	go func() {
		if err := toolBoxServer.Start(); err != nil {
			errChan <- err
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	select {
	case err := <-errChan:
		logger.Error("Toolbox server error", "error", err)
		return 1
	case <-sigChan:
		logger.Info("Received shutdown signal")
		return 0
	}
}
