// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	golog "log"

	"github.com/daytonaio/common-go/pkg/log"
	"github.com/daytonaio/daemon/cmd/daemon/config"
	"github.com/daytonaio/daemon/internal/util"
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

	// Create the console handler with tint for colored output
	consoleHandler := tint.NewHandler(os.Stdout, &tint.Options{
		NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
		TimeFormat: time.RFC3339,
		Level:      logLevel,
	})

	logger := slog.New(consoleHandler)
	slog.SetDefault(logger)

	// Redirect standard library log to slog
	golog.SetOutput(&log.DebugLogWriter{})

	c, err := config.GetConfig()
	if err != nil {
		logger.Error("Failed to get config", "error", err)
		os.Exit(2)
	}

	var logWriter io.Writer
	if c.DaemonLogFilePath != "" {
		logFile, err := os.OpenFile(c.DaemonLogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Error("Failed to open log file", "path", c.DaemonLogFilePath, "error", err)
		} else {
			defer logFile.Close()
			logWriter = logFile

			fileHandler := slog.NewTextHandler(logWriter, &slog.HandlerOptions{
				Level: logLevel,
			})
			handler := log.NewMultiHandler([]slog.Handler{consoleHandler, fileHandler}...)

			logger = slog.New(handler)
			slog.SetDefault(logger)
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed to get user home directory", "error", err)
		return 2
	}

	configDir := filepath.Join(homeDir, ".daytona")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		logger.Error("Failed to create config directory", "path", configDir, "error", err)
		return 2
	}

	// If workdir in image is not set, use user home as workdir
	if c.UserHomeAsWorkDir {
		err = os.Chdir(homeDir)
		if err != nil {
			logger.Warn("Failed to change working directory to home directory", "error", err)
		}
	}

	sessionService := session.NewSessionService(logger, configDir, c.TerminationGracePeriod, c.TerminationCheckInterval)

	// Check if user wants to read entrypoint logs
	args := os.Args[1:]
	if len(args) == 2 && args[0] == "entrypoint" && args[1] == "logs" {
		entrypointLogFilePath := path.Join(configDir, "sessions", util.EntrypointSessionID, util.EntrypointCommandID, "output.log")
		logBytes, err := os.ReadFile(entrypointLogFilePath)
		if err != nil {
			logger.Error("Failed to read entrypoint log file", "error", err)
			fmt.Printf("failed to read entrypoint log file: %v\n", err)
			return 2
		}

		fmt.Print(string(logBytes))
		return 0
	}

	// Execute passed arguments as command in entrypoint session
	if len(args) > 0 {
		// Create entrypoint session
		err = sessionService.Create(util.EntrypointSessionID, false)
		if err != nil {
			logger.Error("Failed to create entrypoint session", "error", err)
			return 2
		}

		logger.Debug("Created entrypoint session", "session_id", util.EntrypointSessionID)

		// Execute command asynchronously via session
		command := util.ShellQuoteJoin(args)
		_, err := sessionService.Execute(
			util.EntrypointSessionID,
			util.EntrypointCommandID,
			command,
			true,  // async=true for non-blocking
			false, // isCombinedOutput=false
			true,  // suppressInputEcho=true
		)
		if err != nil {
			logger.Error("Failed to execute entrypoint command", "error", err)
			return 2
		}
	}

	errChan := make(chan error)

	workDir, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get current working directory", "error", err)
		return 2
	}

	recordingsDir := c.RecordingsDir
	if recordingsDir == "" {
		recordingsDir = filepath.Join(configDir, "recordings")
	}
	recordingService := recording.NewRecordingService(logger, recordingsDir)

	toolBoxServer := toolbox.NewServer(toolbox.ServerConfig{
		Logger:           logger,
		WorkDir:          workDir,
		ConfigDir:        configDir,
		OtelEndpoint:     c.OtelEndpoint,
		SandboxId:        c.SandboxId,
		SessionService:   sessionService,
		RecordingService: recordingService,
	})

	// Start the toolbox server in a go routine
	go func() {
		err := toolBoxServer.Start()
		if err != nil {
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

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either an error or shutdown signal
	select {
	case err := <-errChan:
		logger.Error("Error occurred", "error", err)
	case sig := <-sigChan:
		logger.Info("Received signal, shutting down gracefully...", "signal", sig)
	}

	if len(args) > 0 {
		// Handle entrypoint command shutdown
		_, err = sessionService.Get(util.EntrypointSessionID)
		if err != nil {
			logger.Error("Failed to get entrypoint session", "error", err)
		} else {
			delErr := sessionService.Delete(context.Background(), util.EntrypointSessionID)
			if delErr != nil {
				logger.Error("Failed to delete entrypoint session", "error", delErr)
			}
		}
	}

	// Toolbox server graceful shutdown
	toolBoxServer.Shutdown()

	slog.Info("Shutdown complete")
	return 0
}
