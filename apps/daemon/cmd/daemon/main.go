// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	golog "log"

	"github.com/daytonaio/common-go/pkg/log"
	"github.com/daytonaio/daemon/cmd/daemon/config"
	"github.com/daytonaio/daemon/internal/util"
	"github.com/daytonaio/daemon/pkg/recording"
	"github.com/daytonaio/daemon/pkg/recordingdashboard"
	"github.com/daytonaio/daemon/pkg/ssh"
	"github.com/daytonaio/daemon/pkg/terminal"
	"github.com/daytonaio/daemon/pkg/toolbox"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func main() {
	c, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	// Check if user wants to read entrypoint logs
	args := os.Args[1:]
	if len(args) == 2 && args[0] == "entrypoint" && args[1] == "logs" {
		util.ReadEntrypointLogs(c.EntrypointLogFilePath)
		return
	}

	var logWriter io.Writer
	if c.DaemonLogFilePath != "" {
		logFile, err := os.OpenFile(c.DaemonLogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			slog.Error("Failed to open log file", "path", c.DaemonLogFilePath, "error", err)
		} else {
			defer logFile.Close()
			logWriter = logFile
		}
	}

	initLogs(logWriter)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("failed to get user home directory: %w", err))
	}

	configDir := filepath.Join(homeDir, ".daytona")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		panic(fmt.Errorf("failed to create config directory: %w", err))
	}

	// If workdir in image is not set, use user home as workdir
	if c.UserHomeAsWorkDir {
		err = os.Chdir(homeDir)
		if err != nil {
			slog.Warn("failed to change working directory to home directory", "error", err)
		}
	}

	// Execute passed arguments as command
	var entrypointCmd *exec.Cmd
	var entrypointWg sync.WaitGroup
	if len(args) > 0 {
		// used for logging in case of errors starting/waiting for the command
		entrypointLogWriter := os.Stdout
		entrypointErrLogWriter := os.Stderr

		if c.EntrypointLogFilePath != "" {
			entrypointLogFile, err := os.OpenFile(c.EntrypointLogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				slog.Error("Failed to open log file, fallback to STDOUT and STDERR",
					"path", c.EntrypointLogFilePath,
					"error", err)
			} else {
				defer entrypointLogFile.Close()
				entrypointLogWriter = entrypointLogFile
				entrypointErrLogWriter = entrypointLogFile
			}
		}

		entrypointCmd = exec.Command(args[0], args[1:]...)
		entrypointCmd.Env = os.Environ()
		entrypointCmd.Stdout = entrypointLogWriter
		entrypointCmd.Stderr = entrypointErrLogWriter

		// Start the command and wait for it in a background goroutine.
		// This ensures the child process is properly reaped (preventing zombies)
		// while allowing the daemon to continue initialization without blocking.
		startErr := entrypointCmd.Start()
		if startErr != nil {
			fmt.Fprintf(entrypointErrLogWriter, "failed to start command: %v\n", startErr)
		} else {
			entrypointWg.Add(1)
			go func() {
				defer entrypointWg.Done()
				if err := entrypointCmd.Wait(); err != nil {
					fmt.Fprintf(entrypointErrLogWriter, "command exited with error: %v\n", err)
				} else {
					fmt.Fprint(entrypointLogWriter, "Entrypoint command completed successfully\n")
				}
			}()
		}
	}

	errChan := make(chan error)

	workDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("failed to get current working directory: %w", err))
	}

	recordingsDir := c.RecordingsDir
	if recordingsDir == "" {
		recordingsDir = filepath.Join(configDir, "recordings")
	}
	recordingService := recording.NewRecordingService(recordingsDir)

	toolBoxServer := toolbox.NewServer(toolbox.ServerConfig{
		WorkDir:                              workDir,
		ConfigDir:                            configDir,
		OtelEndpoint:                         c.OtelEndpoint,
		SandboxId:                            c.SandboxId,
		TerminationGracePeriodSeconds:        c.TerminationGracePeriodSeconds,
		TerminationCheckIntervalMilliseconds: c.TerminationCheckIntervalMilliseconds,
		RecordingService:                     recordingService,
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
		if err := recordingdashboard.NewDashboardServer(recordingService).Start(); err != nil {
			errChan <- err
		}
	}()

	sshServer := &ssh.Server{
		WorkDir:        workDir,
		DefaultWorkDir: workDir,
	}
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
		slog.Error("Error", "error", err)
	case sig := <-sigChan:
		slog.Info("Received signal, shutting down gracefully...", "signal", sig)
	}

	// Graceful shutdown
	slog.Info("Stopping computer use processes...")
	if toolBoxServer.ComputerUse != nil {
		_, err := toolBoxServer.ComputerUse.Stop()
		if err != nil {
			slog.Error("Failed to stop computer use", "error", err)
		}
	}

	// Handle entrypoint command shutdown
	if entrypointCmd != nil && entrypointCmd.Process != nil {
		slog.Info("Waiting for entrypoint command to complete...")

		// Create a channel to signal when WaitGroup is done
		done := make(chan struct{})
		go func() {
			entrypointWg.Wait()
			close(done)
		}()

		// Wait with timeout for graceful completion
		timer := time.NewTimer(time.Duration(c.EntrypointShutdownTimeoutSec) * time.Second)
		select {
		case <-done:
			slog.Info("Entrypoint command completed")
			if !timer.Stop() {
				<-timer.C
			}
		case <-timer.C:
			slog.Warn("Entrypoint command did not complete within timeout, sending SIGTERM...")
			if err := entrypointCmd.Process.Signal(syscall.SIGTERM); err != nil {
				slog.Error("Failed to send SIGTERM to entrypoint command", "error", err)
			}

			// Wait a bit more for SIGTERM to take effect
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.SigtermShutdownTimeoutSec)*time.Second)
			defer cancel()

			gracefulDone := make(chan struct{})
			go func() {
				entrypointWg.Wait()
				close(gracefulDone)
			}()

			select {
			case <-gracefulDone:
				slog.Info("Entrypoint command terminated gracefully")
			case <-ctx.Done():
				slog.Warn("Entrypoint command did not respond to SIGTERM, sending SIGKILL...")
				if err := entrypointCmd.Process.Kill(); err != nil {
					slog.Error("Failed to kill entrypoint command", "error", err)
				}
				entrypointWg.Wait()
				slog.Info("Entrypoint command killed")
			}
		}
	}

	slog.Info("Shutdown complete")
}

func initLogs(logWriter io.Writer) {
	logLevel := log.ParseLogLevel(os.Getenv("LOG_LEVEL"))

	// Create the console handler with tint for colored output
	consoleHandler := tint.NewHandler(os.Stdout, &tint.Options{
		NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
		TimeFormat: time.RFC3339,
		Level:      logLevel,
	})

	var handler slog.Handler = consoleHandler

	// If we have a log file writer, create a multi-handler
	if logWriter != nil {
		fileHandler := slog.NewTextHandler(logWriter, &slog.HandlerOptions{
			Level: logLevel,
		})
		handler = log.NewMultiHandler([]slog.Handler{consoleHandler, fileHandler}...)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Redirect standard library log to slog
	golog.SetOutput(&log.DebugLogWriter{})
}
