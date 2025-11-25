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

	common_daemon "github.com/daytonaio/common-go/pkg/daemon"
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

	// If workdir in image is not set, use user home as workdir
	_, userHomeAsWorkDirSet := os.LookupEnv(common_daemon.UserHomeAsWorkDirEnvVar)
	if userHomeAsWorkDirSet {
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
	args := os.Args[1:]
	if len(args) > 0 {
		entrypointCmd = exec.Command(args[0], args[1:]...)
		entrypointCmd.Stdout = &util.PrefixedWriter{Prefix: "[ENTRYPOINT] ", Writer: os.Stdout}
		entrypointCmd.Stderr = &util.PrefixedWriter{Prefix: "[ENTRYPOINT] ", Writer: os.Stderr}

		// Start the command and wait for it in a background goroutine.
		// This ensures the child process is properly reaped (preventing zombies)
		// while allowing the daemon to continue initialization without blocking.
		err := entrypointCmd.Start()
		if err != nil {
			log.Errorf("failed to start command: %v", err)
		} else {
			entrypointWg.Add(1)
			go func() {
				defer entrypointWg.Done()
				if err := entrypointCmd.Wait(); err != nil {
					log.Errorf("command exited with error: %v", err)
				} else {
					log.Info("Entrypoint command completed successfully")
				}
			}()
		}
	}

	var logWriter io.Writer
	if c.LogFilePath != nil {
		logFile, err := os.OpenFile(*c.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Error("Failed to open log file at ", *c.LogFilePath)
		} else {
			defer logFile.Close()
			logWriter = logFile
		}
	}

	initLogs(logWriter)

	errChan := make(chan error)

	workDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("failed to get current working directory: %w", err))
	}

	toolBoxServer := &toolbox.Server{
		WorkDir: workDir,
	}

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
		select {
		case <-done:
			log.Info("Entrypoint command completed")
		case <-time.After(10 * time.Second):
			log.Warn("Entrypoint command did not complete within timeout, sending SIGTERM...")
			if err := entrypointCmd.Process.Signal(syscall.SIGTERM); err != nil {
				log.Errorf("Failed to send SIGTERM to entrypoint command: %v", err)
			}

			// Wait a bit more for SIGTERM to take effect
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
