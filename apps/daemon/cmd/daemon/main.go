// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	golog "log"

	"github.com/daytonaio/daemon/cmd/daemon/config"
	"github.com/daytonaio/daemon/internal/util"
	"github.com/daytonaio/daemon/pkg/ssh"
	"github.com/daytonaio/daemon/pkg/terminal"
	"github.com/daytonaio/daemon/pkg/toolbox"
	log "github.com/sirupsen/logrus"
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
			log.Error("Failed to open log file at ", c.DaemonLogFilePath)
		} else {
			defer logFile.Close()
			logWriter = logFile
		}
	}

	initLogs(logWriter)

	// If workdir in image is not set, use user home as workdir
	if c.UserHomeAsWorkDir {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Warnf("failed to get home directory: %v", err)
		} else {
			err = os.Chdir(homeDir)
			if err != nil {
				log.Warnf("failed to change working directory to home directory: %v", err)
			}
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
				log.Errorf("Failed to open log file at %s due to %v, fallback to STDOUT and STDERR", c.EntrypointLogFilePath, err)
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

	toolBoxServer := toolbox.NewServer(toolbox.ServerConfig{
		WorkDir:      workDir,
		OtelEndpoint: c.OtelEndpoint,
		SandboxId:    c.SandboxId,
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
		log.Errorf("Error: %v", err)
	case sig := <-sigChan:
		log.Infof("Received signal %v, shutting down gracefully...", sig)
	}

	// Graceful shutdown
	log.Info("Stopping computer use processes...")
	if toolBoxServer.ComputerUse != nil {
		_, err := toolBoxServer.ComputerUse.Stop()
		if err != nil {
			log.Errorf("Failed to stop computer use: %v", err)
		}
	}

	// Handle entrypoint command shutdown
	if entrypointCmd != nil && entrypointCmd.Process != nil {
		log.Info("Waiting for entrypoint command to complete...")

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
			log.Info("Entrypoint command completed")
			if !timer.Stop() {
				<-timer.C
			}
		case <-timer.C:
			log.Warn("Entrypoint command did not complete within timeout, sending SIGTERM...")
			if err := entrypointCmd.Process.Signal(syscall.SIGTERM); err != nil {
				log.Errorf("Failed to send SIGTERM to entrypoint command: %v", err)
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
				log.Info("Entrypoint command terminated gracefully")
			case <-ctx.Done():
				log.Warn("Entrypoint command did not respond to SIGTERM, sending SIGKILL...")
				if err := entrypointCmd.Process.Kill(); err != nil {
					log.Errorf("Failed to kill entrypoint command: %v", err)
				}
				entrypointWg.Wait()
				log.Info("Entrypoint command killed")
			}
		}
	}

	log.Info("Shutdown complete")
}

func initLogs(logWriter io.Writer) {
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
	logFormatter := &config.LogFormatter{
		TextFormatter: &log.TextFormatter{
			ForceColors: true,
		},
		LogFileWriter: logWriter,
	}

	log.SetFormatter(logFormatter)

	golog.SetOutput(log.New().WriterLevel(log.DebugLevel))
}
