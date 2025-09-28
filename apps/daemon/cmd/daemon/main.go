// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	golog "log"

	common_daemon "github.com/daytonaio/common-go/pkg/daemon"
	"github.com/daytonaio/daemon/cmd/daemon/config"
	"github.com/daytonaio/daemon/pkg/metrics"
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

	workDirFlag := flag.String("work-dir", "", "optional; sets the working directory; defaults to the current directory; use "+common_daemon.UseUserHomeAsWorkDir+" to switch to the user's home directory")
	flag.Parse()

	if workDirFlag != nil && *workDirFlag != "" {
		workDir := *workDirFlag
		if workDir == common_daemon.UseUserHomeAsWorkDir {
			workDir, err = os.UserHomeDir()
			if err != nil {
				panic(fmt.Errorf("failed to get user home directory: %w", err))
			}
		}
		err = os.Chdir(workDir)
		if err != nil {
			panic(fmt.Errorf("failed to change working directory to %s: %w", workDir, err))
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

	// Get sandbox ID from environment
	sandboxId := os.Getenv("DAYTONA_SANDBOX_ID")
	if sandboxId == "" {
		log.Warn("DAYTONA_SANDBOX_ID environment variable not set")
		sandboxId = "unknown"
	}

	// Initialize and start metrics service
	metricsCollector := metrics.NewCollector("/") // Using root path for disk metrics
	metricsService := metrics.NewService(metricsCollector, 10*time.Second, sandboxId)
	if err := metricsService.Start(); err != nil {
		log.Errorf("Failed to start metrics service: %v", err)
	}

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

	log.Info("Stopping metrics service...")
	metricsService.Stop()

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
