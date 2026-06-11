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
	"github.com/daytonaio/daemon/pkg/recording"
	"github.com/daytonaio/daemon/pkg/recordingdashboard"
	"github.com/daytonaio/daemon/pkg/session"
	"github.com/daytonaio/daemon/pkg/ssh"
	"github.com/daytonaio/daemon/pkg/terminal"
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

	sessionService := session.NewSessionService(logger, configDir, c.TerminationGracePeriod, c.TerminationCheckInterval)

	recordingsDir := c.RecordingsDir
	if recordingsDir == "" {
		recordingsDir = filepath.Join(configDir, "recordings")
	}
	recordingService := recording.NewRecordingService(logger, recordingsDir)

	workDir, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get current working directory", "error", err)
		return 2
	}

	toolBoxServer := toolbox.NewServer(toolbox.ServerConfig{
		Logger:           logger,
		WorkDir:          workDir,
		ConfigDir:        configDir,
		OtelEndpoint:     c.OtelEndpoint,
		SandboxId:        c.SandboxId,
		SessionService:   sessionService,
		RecordingService: recordingService,
		OrganizationId:   c.OrganizationId,
		RegionId:         c.RegionId,
		Snapshot:         c.Snapshot,
	})

	errChan := make(chan error)
	go func() {
		if err := toolBoxServer.Start(); err != nil {
			errChan <- err
		}
	}()

	// Start terminal server
	go func() {
		if err := terminal.StartTerminalServer(22222); err != nil {
			errChan <- err
		}
	}()

	// Start recording dashboard server
	go func() {
		if err := recordingdashboard.NewDashboardServer(logger, recordingService).Start(); err != nil {
			errChan <- err
		}
	}()

	sshServer := ssh.NewServer(logger, workDir, workDir)

	go func() {
		if err := sshServer.Start(); err != nil {
			errChan <- err
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	exitCode := 0
	select {
	case err := <-errChan:
		logger.Error("Server error", "error", err)
		// Unlike Linux main.go (which returns 0 here), keep the pre-port
		// Windows behavior of exiting non-zero on a server error.
		exitCode = 1
	case <-sigChan:
		logger.Info("Received shutdown signal")
	}

	// Toolbox server graceful shutdown
	toolBoxServer.Shutdown()

	slog.Info("Shutdown complete")
	return exitCode
}
