// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	golog "log"

	"github.com/daytonaio/daemon/cmd/daemon/config"
	"github.com/daytonaio/daemon/internal/util"
	"github.com/daytonaio/daemon/pkg/session"
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

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("failed to get user home directory: %w", err))
	}

	configDir := path.Join(homeDir, ".daytona")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		panic(fmt.Errorf("failed to create config directory: %w", err))
	}

	// If workdir in image is not set, use user home as workdir
	if c.UserHomeAsWorkDir {
		err = os.Chdir(homeDir)
		if err != nil {
			log.Warnf("failed to change working directory to home directory: %v", err)
		}
	}

	sessionService := session.NewSessionService(configDir, c.TerminationGracePeriodSeconds, c.TerminationCheckIntervalMilliseconds)

	// Check if user wants to read entrypoint logs
	args := os.Args[1:]
	if len(args) == 2 && args[0] == "entrypoint" && args[1] == "logs" {
		entrypointLogFilePath := path.Join(configDir, "sessions", util.EntrypointSessionID, util.EntrypointCommandID, "output.log")
		logBytes, err := os.ReadFile(entrypointLogFilePath)
		if err != nil {
			log.Errorf("Failed to read entrypoint log file: %v", err)
			os.Exit(1)
		}

		fmt.Print(string(logBytes))
		return
	}

	// Execute passed arguments as command in entrypoint session
	if len(args) > 0 {
		// Create entrypoint session
		err = sessionService.Create(util.EntrypointSessionID, false)
		if err != nil {
			log.Errorf("Failed to create entrypoint session: %v", err)
		} else {
			log.Infof("Created entrypoint session with ID: %s", util.EntrypointSessionID)
		}

		// Execute command asynchronously via session
		command := strings.Join(args, " ")
		_, err := sessionService.Execute(
			util.EntrypointSessionID,
			util.Pointer(util.EntrypointCommandID),
			command,
			true,  // async=true for non-blocking
			false, // isCombinedOutput=false
		)
		if err != nil {
			log.Errorf("Failed to execute entrypoint command: %v", err)
		}
	}

	errChan := make(chan error)

	workDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("failed to get current working directory: %w", err))
	}

	toolBoxServer := &toolbox.Server{
		WorkDir:        workDir,
		ConfigDir:      configDir,
		SessionService: sessionService,
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
	_, err = sessionService.Get(util.EntrypointSessionID)
	if err != nil {
		log.Errorf("Failed to get entrypoint session: %v", err)
	} else {
		delErr := sessionService.Delete(context.Background(), util.EntrypointSessionID)
		if delErr != nil {
			log.Errorf("Failed to delete entrypoint session: %v", delErr)
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
